package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wozai/wozai/internal/model"
)

type mockMsgRepo struct {
	msgs   []*model.Message
	nextID int64
}

func newMockMsgRepo() *mockMsgRepo {
	return &mockMsgRepo{nextID: 1}
}

func (r *mockMsgRepo) Create(_ context.Context, m *model.Message) (*model.Message, error) {
	m.ID = r.nextID
	r.nextID++
	r.msgs = append(r.msgs, m)
	return m, nil
}

func (r *mockMsgRepo) ListBySoulID(_ context.Context, soulID int64, limit int) ([]*model.Message, error) {
	var result []*model.Message
	for _, m := range r.msgs {
		if m.SoulID == soulID {
			result = append(result, m)
		}
	}
	if len(result) > limit {
		result = result[len(result)-limit:]
	}
	return result, nil
}

func TestBuildSystemPrompt(t *testing.T) {
	soul := &model.Soul{
		Name:        "爸爸",
		Relation:    "父亲",
		Personality: "温和",
		SpeechStyle: "慢条斯理",
		Memory:      "一起钓鱼",
	}
	prompt := buildSystemPrompt(soul)
	if len(prompt) == 0 {
		t.Fatal("empty prompt")
	}
	for _, want := range []string{"爸爸", "父亲", "温和", "慢条斯理", "一起钓鱼"} {
		if !containsStr(prompt, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStrImpl(s, sub))
}

func containsStrImpl(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestChat_Success(t *testing.T) {
	// mock DeepSeek server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing auth header")
		}
		resp := dsResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{{Message: struct {
				Content string `json:"content"`
			}{Content: "你好啊，好久不见"}}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	soulRepo := newMockSoulRepo()
	soulRepo.souls[1] = &model.Soul{ID: 1, UserID: 10, Name: "爸爸"}
	msgRepo := newMockMsgRepo()

	svc := NewChatService(msgRepo, soulRepo, srv.URL, "test-key")
	reply, userMsg, asstMsg, err := svc.Chat(context.Background(), 10, 1, "爸，我想你了")
	if err != nil {
		t.Fatal(err)
	}
	if reply != "你好啊，好久不见" {
		t.Errorf("reply = %q", reply)
	}
	if userMsg.Role != "user" || asstMsg.Role != "assistant" {
		t.Error("wrong message roles")
	}
	if len(msgRepo.msgs) != 2 {
		t.Errorf("msgs count = %d; want 2", len(msgRepo.msgs))
	}
}

func TestChat_SoulNotFound(t *testing.T) {
	svc := NewChatService(newMockMsgRepo(), newMockSoulRepo(), "", "")
	_, _, _, err := svc.Chat(context.Background(), 1, 999, "hello")
	if err != ErrSoulNotFound {
		t.Errorf("err = %v; want ErrSoulNotFound", err)
	}
}

func TestChat_SoulForbidden(t *testing.T) {
	soulRepo := newMockSoulRepo()
	soulRepo.souls[1] = &model.Soul{ID: 1, UserID: 10}
	svc := NewChatService(newMockMsgRepo(), soulRepo, "", "")
	_, _, _, err := svc.Chat(context.Background(), 99, 1, "hello")
	if err != ErrSoulForbid {
		t.Errorf("err = %v; want ErrSoulForbid", err)
	}
}

func TestHistory_Success(t *testing.T) {
	soulRepo := newMockSoulRepo()
	soulRepo.souls[1] = &model.Soul{ID: 1, UserID: 10}
	msgRepo := newMockMsgRepo()
	msgRepo.msgs = []*model.Message{
		{ID: 1, SoulID: 1, Role: "user", Content: "hi"},
		{ID: 2, SoulID: 1, Role: "assistant", Content: "hello"},
	}
	svc := NewChatService(msgRepo, soulRepo, "", "")
	msgs, err := svc.History(context.Background(), 10, 1, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 2 {
		t.Errorf("len = %d; want 2", len(msgs))
	}
}
