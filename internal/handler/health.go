package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

// DBPinger checks database connectivity.
type DBPinger interface {
	Ping(ctx context.Context) error
}

// HealthHandler handles health check requests.
type HealthHandler struct {
	db DBPinger
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db DBPinger) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health returns the health status of the application.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	dbStatus := "connected"
	if err := h.db.Ping(r.Context()); err != nil {
		dbStatus = "disconnected"
	}

	w.Header().Set("Content-Type", "application/json")
	status := http.StatusOK
	if dbStatus != "connected" {
		status = http.StatusServiceUnavailable
	}
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(map[string]string{
		"status": func() string {
			if dbStatus == "connected" {
				return "ok"
			}
			return "degraded"
		}(),
		"db": dbStatus,
	})
}
