package handler

import (
	"context"
	"net/http"

	"github.com/wozai/wozai/internal/middleware"
	"github.com/wozai/wozai/internal/model"
)

// ProfileServiceInterface defines profile service methods.
type ProfileServiceInterface interface {
	GetProfile(ctx context.Context, userID int64) (*model.User, error)
	UpdateProfile(ctx context.Context, userID int64, nickname, avatar, bio string) (*model.User, error)
}

// StatsServiceInterface defines stats service methods.
type StatsServiceInterface interface {
	GetStats(ctx context.Context) (*model.ChatStats, error)
	GetUserStats(ctx context.Context, userID int64) (soulsCount, messagesCount int64, err error)
}

// ProfileHandler handles profile-related HTTP requests.
type ProfileHandler struct {
	svc ProfileServiceInterface
}

// NewProfileHandler creates a new ProfileHandler.
func NewProfileHandler(svc ProfileServiceInterface) *ProfileHandler {
	return &ProfileHandler{svc: svc}
}

// GetProfile returns the current user's profile.
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	user, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load profile")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}

// UpdateProfile updates the current user's profile.
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	var req struct {
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
		Bio      string `json:"bio"`
	}
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.svc.UpdateProfile(r.Context(), userID, req.Nickname, req.Avatar, req.Bio)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}

// AdminHandler handles admin-related HTTP requests.
type AdminHandler struct {
	userRepo StatsServiceInterface
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(stats StatsServiceInterface) *AdminHandler {
	return &AdminHandler{userRepo: stats}
}

// Stats returns platform statistics.
func (h *AdminHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.userRepo.GetStats(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load stats")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"stats": stats})
}

// UserStats returns stats for the current user.
func (h *AdminHandler) UserStats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	soulsCount, messagesCount, err := h.userRepo.GetUserStats(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load user stats")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"souls_count":    soulsCount,
		"messages_count": messagesCount,
	})
}
