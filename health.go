package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

type DBPinger interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	db DBPinger
}

func NewHealthHandler(db DBPinger) *HealthHandler {
	return &HealthHandler{db: db}
}

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
