package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockDB struct {
	err error
}

func (m *mockDB) Ping(_ context.Context) error {
	return m.err
}

func TestHealth_OK(t *testing.T) {
	h := NewHealthHandler(&mockDB{})
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	h.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("status = %q, want ok", resp["status"])
	}
	if resp["db"] != "connected" {
		t.Errorf("db = %q, want connected", resp["db"])
	}
}

func TestHealth_DBDown(t *testing.T) {
	h := NewHealthHandler(&mockDB{err: errors.New("db down")})
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	h.Health(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "degraded" {
		t.Errorf("status = %q, want degraded", resp["status"])
	}
}
