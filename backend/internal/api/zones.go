package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

type ZoneHandler struct {
	db    *sql.DB
	redis *redis.Client
}

type Zone struct {
	ID        int    `json:"id"`
	ProjectID int    `json:"project_id"`
	ParentID  *int   `json:"parent_id,omitempty"`
	Name      string `json:"name"`
	Level     int    `json:"level"`
	Path      string `json:"path"`
	Children  []Zone `json:"children,omitempty"`
}

func NewZoneHandler(db *sql.DB, redis *redis.Client) *ZoneHandler {
	return &ZoneHandler{db: db, redis: redis}
}

// GetZoneTree returns hierarchical zone structure
func (h *ZoneHandler) GetZoneTree(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["projectId"]

	// Check cache first
	cacheKey := "zones:" + projectID
	cached, err := h.redis.Get(r.Context(), cacheKey).Result()
	if err == nil && cached != "" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write([]byte(cached))
		return
	}

	// Query all zones for project
	rows, err := h.db.Query(`
		SELECT id, project_id, parent_id, name, level, path
		FROM zones
		WHERE project_id = $1
		ORDER BY path
	`, projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Build flat list
	zones := []Zone{}
	zoneMap := make(map[int]*Zone)

	for rows.Next() {
		var z Zone
		err := rows.Scan(&z.ID, &z.ProjectID, &z.ParentID, &z.Name, &z.Level, &z.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		z.Children = []Zone{}
		zones = append(zones, z)
		zoneMap[z.ID] = &zones[len(zones)-1]
	}

	// Build tree
	var roots []Zone
	for i := range zones {
		if zones[i].ParentID == nil {
			roots = append(roots, zones[i])
		} else {
			parent := zoneMap[*zones[i].ParentID]
			if parent != nil {
				parent.Children = append(parent.Children, zones[i])
			}
		}
	}

	// Cache result
	result, _ := json.Marshal(roots)
	h.redis.Set(r.Context(), cacheKey, string(result), 300) // 5 min cache

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(result)
}

// CreateZone creates a new zone
func (h *ZoneHandler) CreateZone(w http.ResponseWriter, r *http.Request) {
	var zone Zone
	if err := json.NewDecoder(r.Body).Decode(&zone); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Calculate path
	var parentPath string
	if zone.ParentID != nil {
		err := h.db.QueryRow("SELECT path FROM zones WHERE id = $1", *zone.ParentID).Scan(&parentPath)
		if err != nil {
			http.Error(w, "Parent zone not found", http.StatusBadRequest)
			return
		}
	}

	// Insert zone
	var id int
	err := h.db.QueryRow(`
		INSERT INTO zones (project_id, parent_id, name, level, path)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, zone.ProjectID, zone.ParentID, zone.Name, zone.Level, parentPath+"/"+zone.Name).Scan(&id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Clear cache
	h.redis.Del(r.Context(), "zones:"+strconv.Itoa(zone.ProjectID))

	zone.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(zone)
}
