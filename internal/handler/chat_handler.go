package handler

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wozai/wozai/internal/middleware"
	"github.com/wozai/wozai/internal/model"
)

// ChatServiceInterface defines chat service methods.
type ChatServiceInterface interface {
	Chat(ctx context.Context, userID, soulID int64, userMessage string) (string, *model.Message, *model.Message, error)
	History(ctx context.Context, userID, soulID int64, limit int) ([]*model.Message, error)
	ListProviders() []string
}

// TTSServiceInterface defines TTS service methods.
type TTSServiceInterface interface {
	Synthesize(ctx context.Context, text, voice string) (io.ReadCloser, string, error)
}

// ChatHandler handles chat HTTP requests.
type ChatHandler struct {
	chatSvc ChatServiceInterface
	ttsSvc  TTSServiceInterface
}

// NewChatHandler creates a new ChatHandler.
func NewChatHandler(chatSvc ChatServiceInterface, ttsSvc TTSServiceInterface) *ChatHandler {
	return &ChatHandler{chatSvc: chatSvc, ttsSvc: ttsSvc}
}

// Send handles sending a chat message.
func (h *ChatHandler) Send(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	soulID, err := strconv.ParseInt(chi.URLParam(r, "soulID"), 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid soul ID")
		return
	}

	var req struct {
		Message string `json:"message"`
	}
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Message == "" {
		WriteError(w, http.StatusBadRequest, "message is required")
		return
	}

	reply, userMsg, assistantMsg, err := h.chatSvc.Chat(r.Context(), userID, soulID, req.Message)
	if err != nil {
		handleSoulError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"reply":         reply,
		"user_msg":      userMsg,
		"assistant_msg": assistantMsg,
	})
}

// History handles retrieving chat history.
func (h *ChatHandler) History(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	soulID, err := strconv.ParseInt(chi.URLParam(r, "soulID"), 10, 64)
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

	msgs, err := h.chatSvc.History(r.Context(), userID, soulID, limit)
	if err != nil {
		handleSoulError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{"messages": msgs})
}

// Speak handles TTS synthesis.
func (h *ChatHandler) Speak(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text  string `json:"text"`
		Voice string `json:"voice"`
	}
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	body, contentType, err := h.ttsSvc.Synthesize(r.Context(), req.Text, req.Voice)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "speech synthesis failed")
		return
	}
	defer body.Close()

	w.Header().Set("Content-Type", contentType)
	io.Copy(w, body)
}

// Providers lists available AI providers.
func (h *ChatHandler) Providers(w http.ResponseWriter, r *http.Request) {
	providers := h.chatSvc.ListProviders()
	WriteJSON(w, http.StatusOK, map[string]interface{}{"providers": providers})
}
