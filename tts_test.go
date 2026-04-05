package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTTS_Synthesize_Success(t *testing.T) {
	audioData := []byte("fake-audio-data")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing auth header")
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Write(audioData)
	}))
	defer srv.Close()

	svc := NewTTSService(srv.URL, "test-key")
	body, ct, err := svc.Synthesize(context.Background(), "你好", "")
	if err != nil {
		t.Fatal(err)
	}
	defer body.Close()

	if ct != "audio/mpeg" {
		t.Errorf("content-type = %q; want audio/mpeg", ct)
	}

	data, _ := io.ReadAll(body)
	if string(data) != string(audioData) {
		t.Errorf("body = %q; want %q", data, audioData)
	}
}

func TestTTS_Synthesize_EmptyText(t *testing.T) {
	svc := NewTTSService("http://localhost", "key")
	_, _, err := svc.Synthesize(context.Background(), "", "")
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestTTS_Synthesize_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer srv.Close()

	svc := NewTTSService(srv.URL, "key")
	_, _, err := svc.Synthesize(context.Background(), "你好", "")
	if err == nil {
		t.Error("expected error for 429 status")
	}
}
