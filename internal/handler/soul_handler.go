package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wozai/wozai/internal/middleware"
	"github.com/wozai/wozai/internal/model"
	"github.com/wozai/wozai/internal/service"
)

// SoulServiceInterface defines soul service methods.
type SoulServiceInterface interface {
	Create(ctx context.Context, userID int64, name, relation, personality, speechStyle, memory string) (*model.Soul, error)
	Get(ctx context.Context, userID, soulID int64) (*model.Soul, error)
	List(ctx context.Context, userID int64) ([]*model.Soul, error)
	Update(ctx context.Context, userID, soulID int64, name, relation, personality, speechStyle, memory string) (*model.Soul, error)
	Delete(ctx context.Context, userID, soulID int64) error
	History(ctx context.Context, userID, soulID int64, limit int) ([]*model.SoulHistory, error)
}

// SoulHandler handles soul HTTP requests.
type SoulHandler struct {
	svc   SoulServiceInterface
	audit *middleware.AuditLogger
}

// NewSoulHandler creates a new SoulHandler.
func NewSoulHandler(svc SoulServiceInterface, audit *middleware.AuditLogger) *SoulHandler {
	return &SoulHandler{svc: svc, audit: audit}
}

// Create handles soul creation.
func (h *SoulHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	var req struct {
		Name        string `json:"name"`
		Relation    string `json:"relation"`
		Personality string `json:"personality"`
		SpeechStyle string `json:"speech_style"`
		Memory      string `json:"memory"`
	}
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	soul, err := h.svc.Create(r.Context(), userID, req.Name, req.Relation, req.Personality, req.SpeechStyle, req.Memory)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSoulName):
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrSoulLimit):
			WriteError(w, http.StatusForbidden, err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, "create soul failed")
		}
		return
	}

	if h.audit != nil {
		h.audit.Log(r.Context(), userID, "create", "soul", soul.Name, r.RemoteAddr)
	}

	WriteJSON(w, http.StatusCreated, map[string]interface{}{"soul": soul})
}

// Get handles soul retrieval.
func (h *SoulHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	soulID, err := parseSoulID(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid soul ID")
		return
	}

	soul, err := h.svc.Get(r.Context(), userID, soulID)
	if err != nil {
		handleSoulError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{"soul": soul})
}

// List handles soul listing.
func (h *SoulHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	souls, err := h.svc.List(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "list souls failed")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{"souls": souls})
}

// Update handles soul update.
func (h *SoulHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	soulID, err := parseSoulID(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid soul ID")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Relation    string `json:"relation"`
		Personality string `json:"personality"`
		SpeechStyle string `json:"speech_style"`
		Memory      string `json:"memory"`
	}
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	soul, err := h.svc.Update(r.Context(), userID, soulID, req.Name, req.Relation, req.Personality, req.SpeechStyle, req.Memory)
	if err != nil {
		handleSoulError(w, err)
		return
	}

	if h.audit != nil {
		h.audit.Log(r.Context(), userID, "update", "soul", soul.Name, r.RemoteAddr)
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{"soul": soul})
}

// Delete handles soul deletion.
func (h *SoulHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	soulID, err := parseSoulID(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid soul ID")
		return
	}

	if err := h.svc.Delete(r.Context(), userID, soulID); err != nil {
		handleSoulError(w, err)
		return
	}

	if h.audit != nil {
		h.audit.Log(r.Context(), userID, "delete", "soul", "", r.RemoteAddr)
	}

	w.WriteHeader(http.StatusNoContent)
}

// EditHistory returns the edit history for a soul.
func (h *SoulHandler) EditHistory(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	soulID, err := parseSoulID(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid soul ID")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	history, err := h.svc.History(r.Context(), userID, soulID, limit)
	if err != nil {
		handleSoulError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{"history": history})
}

func parseSoulID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "soulID"), 10, 64)
}

func handleSoulError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrSoulNotFound):
		WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrSoulForbid):
		WriteError(w, http.StatusForbidden, err.Error())
	default:
		WriteError(w, http.StatusInternalServerError, "operation failed")
	}
}
