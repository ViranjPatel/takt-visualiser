package api

import (
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

// SetupRoutes configures all API endpoints
func SetupRoutes(r *mux.Router, db *sql.DB, redis *redis.Client) {
	// API version prefix
	api := r.PathPrefix("/api/v1").Subrouter()

	// Health check
	api.HandleFunc("/health", HealthHandler).Methods("GET")

	// Tasks endpoints
	taskHandler := NewTaskHandler(db, redis)
	api.HandleFunc("/tasks", taskHandler.GetTasks).Methods("GET")
	api.HandleFunc("/tasks/bulk", taskHandler.BulkUpdate).Methods("PATCH")
	api.HandleFunc("/tasks/{id}", taskHandler.GetTask).Methods("GET")
	api.HandleFunc("/tasks/{id}", taskHandler.UpdateTask).Methods("PATCH")

	// Zones endpoints
	zoneHandler := NewZoneHandler(db, redis)
	api.HandleFunc("/zones/{projectId}/tree", zoneHandler.GetZoneTree).Methods("GET")
	api.HandleFunc("/zones", zoneHandler.CreateZone).Methods("POST")

	// WebSocket for real-time updates
	api.HandleFunc("/ws", NewWebSocketHandler(redis).Handle)
}
