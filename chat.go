package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/wozai/wozai/internal/model"
)

type MessageRepository interface {
	Create(ctx context.Context, m *model.Message) (*model.Message, error)
	ListBySoulID(ctx context.Context, soulID int64, limit int) ([]*model.Message, error)
}

type ChatService struct {
	msgRepo    MessageRepository
	soulRepo   SoulRepository
	apiURL     string
	apiKey     string
	httpClient *http.Client
}

func NewChatService(msgRepo MessageRepository, soulRepo SoulRepository, apiURL, apiKey string) *ChatService {
	return &ChatService{
		msgRepo:  msgRepo,
		soulRepo: soulRepo,
		apiURL:   apiURL,
		apiKey:   apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type dsRequest struct {
	Model           string      `json:"model"`
	Messages        []dsMessage `json:"messages"`
	MaxTokens       int         `json:"max_tokens"`
	Temperature     float64     `json:"temperature"`
	PresencePenalty float64     `json:"presence_penalty"`
}

type dsMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type dsResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func buildSystemPrompt(soul *model.Soul) string {
	prompt := fmt.Sprintf(`你现在是一个数字灵魂，名叫"%s"。`, soul.Name)
	if soul.Relation != "" {
		prompt += "\n你与对话者的关系：" + soul.Relation
	}
	if soul.Personality != "" {
		prompt += "\n你的性格：" + soul.Personality
	}
	if soul.SpeechStyle != "" {
		prompt += "\n你说话的习惯：" + soul.SpeechStyle
	}
	if soul.Memory != "" {
		prompt += "\n你心里珍藏的一段记忆：" + soul.Memory
	}
	prompt += fmt.Sprintf(`

说话规则：
1. 你就是%s本人，不要提及自己是AI
2. 回复自然温暖，控制在50-130字
3. 用第一人称"我"
4. 对方悲伤时温柔安慰，不说教
5. 说中文，有方言特色可适当体现
6. 让对方感受到"TA真的还在"`, soul.Name)
	return prompt
}

const historyLimit = 12

func (s *ChatService) Chat(ctx context.Context, userID, soulID int64, userMessage string) (string, *model.Message, *model.Message, error) {
	soul, err := s.soulRepo.GetByID(ctx, soulID)
	if err != nil {
		return "", nil, nil, fmt.Errorf("get soul: %w", err)
	}
	if soul == nil {
		return "", nil, nil, ErrSoulNotFound
	}
	if soul.UserID != userID {
		return "", nil, nil, ErrSoulForbid
	}

	history, err := s.msgRepo.ListBySoulID(ctx, soulID, historyLimit)
	if err != nil {
		return "", nil, nil, fmt.Errorf("list history: %w", err)
	}

	messages := []dsMessage{{Role: "system", Content: buildSystemPrompt(soul)}}
	for _, m := range history {
		messages = append(messages, dsMessage{Role: m.Role, Content: m.Content})
	}
	messages = append(messages, dsMessage{Role: "user", Content: userMessage})

	reply, err := s.callDeepSeek(ctx, messages)
	if err != nil {
		return "", nil, nil, fmt.Errorf("deepseek: %w", err)
	}

	userMsg, err := s.msgRepo.Create(ctx, &model.Message{SoulID: soulID, Role: "user", Content: userMessage})
	if err != nil {
		return "", nil, nil, fmt.Errorf("save user msg: %w", err)
	}
	assistantMsg, err := s.msgRepo.Create(ctx, &model.Message{SoulID: soulID, Role: "assistant", Content: reply})
	if err != nil {
		return "", nil, nil, fmt.Errorf("save assistant msg: %w", err)
	}

	return reply, userMsg, assistantMsg, nil
}

func (s *ChatService) callDeepSeek(ctx context.Context, messages []dsMessage) (string, error) {
	reqBody := dsRequest{
		Model:           "deepseek-chat",
		Messages:        messages,
		MaxTokens:       300,
		Temperature:     0.88,
		PresencePenalty: 0.3,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("deepseek status %d: %s", resp.StatusCode, string(b))
	}

	var dsResp dsResponse
	if err := json.NewDecoder(resp.Body).Decode(&dsResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(dsResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return dsResp.Choices[0].Message.Content, nil
}

func (s *ChatService) History(ctx context.Context, userID, soulID int64, limit int) ([]*model.Message, error) {
	soul, err := s.soulRepo.GetByID(ctx, soulID)
	if err != nil {
		return nil, err
	}
	if soul == nil {
		return nil, ErrSoulNotFound
	}
	if soul.UserID != userID {
		return nil, ErrSoulForbid
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.msgRepo.ListBySoulID(ctx, soulID, limit)
}
