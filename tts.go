package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TTSService struct {
	apiURL     string
	apiKey     string
	httpClient *http.Client
}

func NewTTSService(apiURL, apiKey string) *TTSService {
	return &TTSService{
		apiURL: apiURL,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type ttsRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
	Voice string `json:"voice"`
}

func (s *TTSService) Synthesize(ctx context.Context, text, voice string) (io.ReadCloser, string, error) {
	if text == "" {
		return nil, "", fmt.Errorf("text is empty")
	}
	if voice == "" {
		voice = "male_1"
	}

	reqBody := ttsRequest{
		Model: "FishAudio/fish-speech-1.5",
		Input: text,
		Voice: voice,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, "", fmt.Errorf("siliconflow status %d: %s", resp.StatusCode, string(b))
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "audio/mpeg"
	}

	return resp.Body, contentType, nil
}
