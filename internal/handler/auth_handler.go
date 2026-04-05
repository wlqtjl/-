package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/wozai/wozai/internal/middleware"
	"github.com/wozai/wozai/internal/model"
	"github.com/wozai/wozai/internal/service"
)

// AuthServiceInterface defines auth service methods.
type AuthServiceInterface interface {
	Register(ctx context.Context, email, password, nickname string) (*model.User, *service.TokenPair, error)
	Login(ctx context.Context, email, password string) (*model.User, *service.TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*service.TokenPair, error)
}

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	svc   AuthServiceInterface
	audit *middleware.AuditLogger
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(svc AuthServiceInterface, audit *middleware.AuditLogger) *AuthHandler {
	return &AuthHandler{svc: svc, audit: audit}
}

// Register handles user registration.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
	}
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, tokens, err := h.svc.Register(r.Context(), req.Email, req.Password, req.Nickname)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailTaken):
			WriteError(w, http.StatusConflict, err.Error())
		case errors.Is(err, service.ErrWeakPassword):
			WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrInvalidEmail):
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, "registration failed")
		}
		return
	}

	if h.audit != nil {
		h.audit.Log(r.Context(), user.ID, "register", "user", "", r.RemoteAddr)
	}

	WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"user":   user,
		"tokens": tokens,
	})
}

// Login handles user login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, tokens, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCreds) {
			WriteError(w, http.StatusUnauthorized, err.Error())
		} else {
			WriteError(w, http.StatusInternalServerError, "login failed")
		}
		return
	}

	if h.audit != nil {
		h.audit.Log(r.Context(), user.ID, "login", "user", "", r.RemoteAddr)
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"user":   user,
		"tokens": tokens,
	})
}

// Refresh handles token refresh.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tokens, err := h.svc.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	WriteJSON(w, http.StatusOK, tokens)
}
