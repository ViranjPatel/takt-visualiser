package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

type TaskHandler struct {
	db    *sql.DB
	redis *redis.Client
}

type Task struct {
	ID             int       `json:"id"`
	ProjectID      int       `json:"project_id"`
	ZoneID         int       `json:"zone_id"`
	Name           string    `json:"name"`
	StartDate      string    `json:"start_date"`
	Duration       int       `json:"duration"`
	TradeID        *int      `json:"trade_id,omitempty"`
	Status         string    `json:"status"`
	SequenceNumber *int      `json:"sequence_number,omitempty"`
	Color          *string   `json:"color,omitempty"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func NewTaskHandler(db *sql.DB, redis *redis.Client) *TaskHandler {
	return &TaskHandler{db: db, redis: redis}
}

// GetTasks returns tasks filtered by query params
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Parse query params
	projectID := r.URL.Query().Get("project_id")
	zoneIDs := r.URL.Query().Get("zone_ids")
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	// Build query
	query := `SELECT id, project_id, zone_id, name, start_date, duration, 
	          trade_id, status, sequence_number, color, updated_at 
	          FROM tasks WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	if projectID != "" {
		query += " AND project_id = $" + strconv.Itoa(argIndex)
		args = append(args, projectID)
		argIndex++
	}

	if zoneIDs != "" {
		zoneList := strings.Split(zoneIDs, ",")
		placeholders := make([]string, len(zoneList))
		for i, id := range zoneList {
			placeholders[i] = "$" + strconv.Itoa(argIndex)
			args = append(args, id)
			argIndex++
		}
		query += " AND zone_id IN (" + strings.Join(placeholders, ",") + ")"
	}

	if dateFrom != "" {
		query += " AND start_date >= $" + strconv.Itoa(argIndex)
		args = append(args, dateFrom)
		argIndex++
	}

	if dateTo != "" {
		query += " AND start_date <= $" + strconv.Itoa(argIndex)
		args = append(args, dateTo)
	}

	query += " ORDER BY start_date, sequence_number"

	// Execute query
	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.ProjectID, &t.ZoneID, &t.Name, &t.StartDate,
			&t.Duration, &t.TradeID, &t.Status, &t.SequenceNumber, &t.Color, &t.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, t)
	}

	// Log response time
	elapsed := time.Since(start)
	w.Header().Set("X-Response-Time", elapsed.String())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// BulkUpdate handles batch task updates
func (h *TaskHandler) BulkUpdate(w http.ResponseWriter, r *http.Request) {
	var updates []map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	updatedIDs := []int{}
	for _, update := range updates {
		id, ok := update["id"].(float64)
		if !ok {
			continue
		}

		// Build dynamic update query
		setClauses := []string{}
		args := []interface{}{}
		argIndex := 1

		for field, value := range update {
			if field != "id" {
				setClauses = append(setClauses, field+" = $"+strconv.Itoa(argIndex))
				args = append(args, value)
				argIndex++
			}
		}

		if len(setClauses) > 0 {
			args = append(args, int(id))
			query := "UPDATE tasks SET " + strings.Join(setClauses, ", ") +
				" WHERE id = $" + strconv.Itoa(argIndex)
			_, err := tx.Exec(query, args...)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			updatedIDs = append(updatedIDs, int(id))
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Publish updates via Redis
	for _, id := range updatedIDs {
		h.redis.Publish(r.Context(), "task_updates", id)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"updated": len(updatedIDs),
		"ids":     updatedIDs,
	})
}

// GetTask returns a single task
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var t Task
	err := h.db.QueryRow(`
		SELECT id, project_id, zone_id, name, start_date, duration,
		       trade_id, status, sequence_number, color, updated_at
		FROM tasks WHERE id = $1
	`, id).Scan(&t.ID, &t.ProjectID, &t.ZoneID, &t.Name, &t.StartDate,
		&t.Duration, &t.TradeID, &t.Status, &t.SequenceNumber, &t.Color, &t.UpdatedAt)

	if err == sql.ErrNoRows {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// UpdateTask updates a single task
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var update map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Build update query
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	for field, value := range update {
		setClauses = append(setClauses, field+" = $"+strconv.Itoa(argIndex))
		args = append(args, value)
		argIndex++
	}

	args = append(args, id)
	query := "UPDATE tasks SET " + strings.Join(setClauses, ", ") +
		" WHERE id = $" + strconv.Itoa(argIndex)

	_, err := h.db.Exec(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Publish update
	h.redis.Publish(r.Context(), "task_updates", id)

	w.WriteHeader(http.StatusNoContent)
}
