package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTTS_Synthesize_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing auth header")
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Write([]byte("fake-audio-data"))
	}))
	defer server.Close()

	svc := NewTTSService(server.URL, "test-key")
	body, contentType, err := svc.Synthesize(context.Background(), "你好", "male_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer body.Close()

	if contentType != "audio/mpeg" {
		t.Errorf("contentType = %q", contentType)
	}

	data, _ := io.ReadAll(body)
	if string(data) != "fake-audio-data" {
		t.Errorf("unexpected audio data: %q", string(data))
	}
}

func TestTTS_Synthesize_EmptyText(t *testing.T) {
	svc := NewTTSService("http://localhost", "key")
	_, _, err := svc.Synthesize(context.Background(), "", "")
	if err == nil {
		t.Fatal("expected error for empty text")
	}
}

func TestTTS_Synthesize_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("rate limited"))
	}))
	defer server.Close()

	svc := NewTTSService(server.URL, "key")
	_, _, err := svc.Synthesize(context.Background(), "test", "")
	if err == nil {
		t.Fatal("expected error for API error")
	}
}
