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

// mockAIProvider for testing
type mockAIProvider struct {
	reply string
	err   error
}

func (p *mockAIProvider) Name() string { return "mock" }
func (p *mockAIProvider) ChatCompletion(_ context.Context, _ []ChatMessage) (string, error) {
	return p.reply, p.err
}

func TestBuildSystemPrompt(t *testing.T) {
	soul := &model.Soul{
		Name:        "爸爸",
		Relation:    "父亲",
		Personality: "温柔善良",
		SpeechStyle: "慢慢说话",
		Memory:      "每次考试回家都在门口等",
	}
	prompt := BuildSystemPrompt(soul)
	for _, s := range []string{"爸爸", "父亲", "温柔善良", "慢慢说话", "每次考试"} {
		if !containsStr(prompt, s) {
			t.Errorf("prompt missing %q", s)
		}
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && containsStrImpl(s, sub)
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
	soulRepo := newMockSoulRepo()
	soulRepo.souls[1] = &model.Soul{ID: 1, UserID: 10, Name: "爸爸"}
	msgRepo := newMockMsgRepo()

	// Mock AI provider
	provider := &mockAIProvider{reply: "你好孩子"}
	providers := map[string]AIProvider{"mock": provider}

	svc := NewChatService(msgRepo, soulRepo, providers, "mock", false)

	reply, userMsg, assistantMsg, err := svc.Chat(context.Background(), 10, 1, "爸爸你好")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply != "你好孩子" {
		t.Errorf("reply = %q", reply)
	}
	if userMsg == nil || assistantMsg == nil {
		t.Fatal("messages are nil")
	}
	if userMsg.Role != "user" || assistantMsg.Role != "assistant" {
		t.Error("wrong message roles")
	}
}

func TestChat_WithDeepSeekProvider(t *testing.T) {
	// Mock DeepSeek HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing or wrong auth header")
		}
		json.NewEncoder(w).Encode(dsResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{{Message: struct {
				Content string `json:"content"`
			}{Content: "测试回复"}}},
		})
	}))
	defer server.Close()

	soulRepo := newMockSoulRepo()
	soulRepo.souls[1] = &model.Soul{ID: 1, UserID: 10, Name: "爸爸"}
	msgRepo := newMockMsgRepo()

	provider := NewDeepSeekProvider(server.URL, "test-key")
	providers := map[string]AIProvider{"deepseek": provider}

	svc := NewChatService(msgRepo, soulRepo, providers, "deepseek", false)

	reply, _, _, err := svc.Chat(context.Background(), 10, 1, "你好")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply != "测试回复" {
		t.Errorf("reply = %q", reply)
	}
}

func TestChat_WithZhipuProvider(t *testing.T) {
	// Mock Zhipu HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer zhipu-test-key" {
			t.Error("missing or wrong auth header")
		}
		json.NewEncoder(w).Encode(dsResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{{Message: struct {
				Content string `json:"content"`
			}{Content: "智谱回复"}}},
		})
	}))
	defer server.Close()

	soulRepo := newMockSoulRepo()
	soulRepo.souls[1] = &model.Soul{ID: 1, UserID: 10, Name: "爸爸"}
	msgRepo := newMockMsgRepo()

	provider := NewZhipuProvider(server.URL, "zhipu-test-key")
	providers := map[string]AIProvider{"zhipu": provider}

	svc := NewChatService(msgRepo, soulRepo, providers, "zhipu", false)

	reply, _, _, err := svc.Chat(context.Background(), 10, 1, "你好")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply != "智谱回复" {
		t.Errorf("reply = %q, want %q", reply, "智谱回复")
	}
}

func TestChat_SoulNotFound(t *testing.T) {
	soulRepo := newMockSoulRepo()
	msgRepo := newMockMsgRepo()
	provider := &mockAIProvider{reply: ""}
	providers := map[string]AIProvider{"mock": provider}

	svc := NewChatService(msgRepo, soulRepo, providers, "mock", false)

	_, _, _, err := svc.Chat(context.Background(), 10, 999, "hello")
	if err != ErrSoulNotFound {
		t.Fatalf("expected ErrSoulNotFound, got %v", err)
	}
}

func TestChat_SoulForbidden(t *testing.T) {
	soulRepo := newMockSoulRepo()
	soulRepo.souls[1] = &model.Soul{ID: 1, UserID: 10}
	msgRepo := newMockMsgRepo()
	provider := &mockAIProvider{reply: ""}
	providers := map[string]AIProvider{"mock": provider}

	svc := NewChatService(msgRepo, soulRepo, providers, "mock", false)

	_, _, _, err := svc.Chat(context.Background(), 999, 1, "hello")
	if err != ErrSoulForbid {
		t.Fatalf("expected ErrSoulForbid, got %v", err)
	}
}

func TestHistory_Success(t *testing.T) {
	soulRepo := newMockSoulRepo()
	soulRepo.souls[1] = &model.Soul{ID: 1, UserID: 10}
	msgRepo := newMockMsgRepo()
	msgRepo.msgs = []*model.Message{
		{ID: 1, SoulID: 1, Role: "user", Content: "你好"},
		{ID: 2, SoulID: 1, Role: "assistant", Content: "你好孩子"},
	}
	provider := &mockAIProvider{reply: ""}
	providers := map[string]AIProvider{"mock": provider}

	svc := NewChatService(msgRepo, soulRepo, providers, "mock", false)

	msgs, err := svc.History(context.Background(), 10, 1, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}
}

func TestSentimentAnalysis(t *testing.T) {
	tests := []struct {
		text     string
		expected string
	}{
		{"我很想念你", "sad"},
		{"今天很开心", "happy"},
		{"真的很讨厌", "angry"},
		{"今天天气不错", "neutral"},
	}
	for _, tt := range tests {
		result := analyzeSentiment(tt.text)
		if result != tt.expected {
			t.Errorf("analyzeSentiment(%q) = %q, want %q", tt.text, result, tt.expected)
		}
	}
}
