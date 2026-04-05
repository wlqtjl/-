package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/wozai/wozai/internal/model"
)

// MessageRepository defines data access methods for messages.
type MessageRepository interface {
	Create(ctx context.Context, m *model.Message) (*model.Message, error)
	ListBySoulID(ctx context.Context, soulID int64, limit int) ([]*model.Message, error)
}

// AIProvider abstracts different AI chat backends.
type AIProvider interface {
	ChatCompletion(ctx context.Context, messages []ChatMessage) (string, error)
	Name() string
}

// ChatMessage is a generic chat message for AI providers.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatService handles chat interactions.
type ChatService struct {
	msgRepo         MessageRepository
	soulRepo        SoulRepository
	defaultProvider AIProvider
	providers       map[string]AIProvider
	enableSentiment bool
}

// NewChatService creates a new ChatService with multi-model support.
func NewChatService(msgRepo MessageRepository, soulRepo SoulRepository, providers map[string]AIProvider, defaultProvider string, enableSentiment bool) *ChatService {
	cs := &ChatService{
		msgRepo:         msgRepo,
		soulRepo:        soulRepo,
		providers:       providers,
		enableSentiment: enableSentiment,
	}
	if p, ok := providers[defaultProvider]; ok {
		cs.defaultProvider = p
	}
	// fallback: use first provider found
	if cs.defaultProvider == nil {
		for _, p := range providers {
			cs.defaultProvider = p
			break
		}
	}
	return cs
}

// DeepSeekProvider implements AIProvider for DeepSeek.
type DeepSeekProvider struct {
	apiURL     string
	apiKey     string
	httpClient *http.Client
}

// NewDeepSeekProvider creates a DeepSeek AI provider.
func NewDeepSeekProvider(apiURL, apiKey string) *DeepSeekProvider {
	return &DeepSeekProvider{
		apiURL: apiURL,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *DeepSeekProvider) Name() string { return "deepseek" }

func (p *DeepSeekProvider) ChatCompletion(ctx context.Context, messages []ChatMessage) (string, error) {
	reqBody := dsRequest{
		Model:           "deepseek-chat",
		Messages:        toDSMessages(messages),
		MaxTokens:       300,
		Temperature:     0.88,
		PresencePenalty: 0.3,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
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

// OpenAIProvider implements AIProvider for OpenAI-compatible APIs.
type OpenAIProvider struct {
	apiURL     string
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewOpenAIProvider creates an OpenAI-compatible AI provider.
func NewOpenAIProvider(apiURL, apiKey, model string) *OpenAIProvider {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIProvider{
		apiURL: apiURL,
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *OpenAIProvider) Name() string { return "openai" }

func (p *OpenAIProvider) ChatCompletion(ctx context.Context, messages []ChatMessage) (string, error) {
	reqBody := dsRequest{
		Model:           p.model,
		Messages:        toDSMessages(messages),
		MaxTokens:       300,
		Temperature:     0.88,
		PresencePenalty: 0.3,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("openai status %d: %s", resp.StatusCode, string(b))
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

// ZhipuProvider implements AIProvider for Zhipu (ChatGLM) APIs.
type ZhipuProvider struct {
	apiURL     string
	apiKey     string
	httpClient *http.Client
}

// NewZhipuProvider creates a Zhipu AI provider.
func NewZhipuProvider(apiURL, apiKey string) *ZhipuProvider {
	return &ZhipuProvider{
		apiURL: apiURL,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *ZhipuProvider) Name() string { return "zhipu" }

func (p *ZhipuProvider) ChatCompletion(ctx context.Context, messages []ChatMessage) (string, error) {
	reqBody := dsRequest{
		Model:           "glm-4-flash",
		Messages:        toDSMessages(messages),
		MaxTokens:       300,
		Temperature:     0.88,
		PresencePenalty: 0.3,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("zhipu status %d: %s", resp.StatusCode, string(b))
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

func toDSMessages(msgs []ChatMessage) []dsMessage {
	result := make([]dsMessage, len(msgs))
	for i, m := range msgs {
		result[i] = dsMessage{Role: m.Role, Content: m.Content}
	}
	return result
}

// BuildSystemPrompt constructs a personality prompt for the soul.
func BuildSystemPrompt(soul *model.Soul) string {
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

// Chat sends a message and gets an AI reply.
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

	messages := []ChatMessage{{Role: "system", Content: BuildSystemPrompt(soul)}}
	for _, m := range history {
		messages = append(messages, ChatMessage{Role: m.Role, Content: m.Content})
	}
	messages = append(messages, ChatMessage{Role: "user", Content: userMessage})

	reply, err := s.defaultProvider.ChatCompletion(ctx, messages)
	if err != nil {
		return "", nil, nil, fmt.Errorf("ai provider: %w", err)
	}

	// Basic sentiment analysis
	sentiment := ""
	if s.enableSentiment {
		sentiment = analyzeSentiment(userMessage)
	}

	userMsg, err := s.msgRepo.Create(ctx, &model.Message{SoulID: soulID, Role: "user", Content: userMessage, Sentiment: sentiment})
	if err != nil {
		return "", nil, nil, fmt.Errorf("save user msg: %w", err)
	}
	assistantMsg, err := s.msgRepo.Create(ctx, &model.Message{SoulID: soulID, Role: "assistant", Content: reply})
	if err != nil {
		return "", nil, nil, fmt.Errorf("save assistant msg: %w", err)
	}

	return reply, userMsg, assistantMsg, nil
}

// History returns message history for a soul.
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

// ListProviders returns available AI providers.
func (s *ChatService) ListProviders() []string {
	names := make([]string, 0, len(s.providers))
	for name := range s.providers {
		names = append(names, name)
	}
	return names
}

// analyzeSentiment performs basic keyword-based sentiment analysis on Chinese text.
func analyzeSentiment(text string) string {
	sadWords := []string{"想念", "思念", "难过", "伤心", "哭", "泪", "痛", "离开", "不在", "失去", "悲"}
	happyWords := []string{"开心", "快乐", "高兴", "幸福", "笑", "好", "棒", "感谢", "谢谢", "爱"}
	angryWords := []string{"生气", "愤怒", "讨厌", "烦", "恨", "不公平"}

	sadCount, happyCount, angryCount := 0, 0, 0
	for _, w := range sadWords {
		if strings.Contains(text, w) {
			sadCount++
		}
	}
	for _, w := range happyWords {
		if strings.Contains(text, w) {
			happyCount++
		}
	}
	for _, w := range angryWords {
		if strings.Contains(text, w) {
			angryCount++
		}
	}

	if sadCount > happyCount && sadCount > angryCount {
		return "sad"
	}
	if happyCount > sadCount && happyCount > angryCount {
		return "happy"
	}
	if angryCount > 0 {
		return "angry"
	}
	return "neutral"
}
