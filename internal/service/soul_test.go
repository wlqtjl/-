package service

import (
	"context"
	"testing"

	"github.com/wozai/wozai/internal/model"
)

type mockSoulRepo struct {
	souls  map[int64]*model.Soul
	nextID int64
}

func newMockSoulRepo() *mockSoulRepo {
	return &mockSoulRepo{souls: make(map[int64]*model.Soul), nextID: 1}
}

func (m *mockSoulRepo) Create(_ context.Context, s *model.Soul) (*model.Soul, error) {
	s.ID = m.nextID
	m.nextID++
	m.souls[s.ID] = s
	return s, nil
}

func (m *mockSoulRepo) GetByID(_ context.Context, id int64) (*model.Soul, error) {
	s, ok := m.souls[id]
	if !ok {
		return nil, nil
	}
	return s, nil
}

func (m *mockSoulRepo) ListByUserID(_ context.Context, userID int64) ([]*model.Soul, error) {
	var result []*model.Soul
	for _, s := range m.souls {
		if s.UserID == userID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *mockSoulRepo) Update(_ context.Context, s *model.Soul) (*model.Soul, error) {
	existing, ok := m.souls[s.ID]
	if !ok || existing.UserID != s.UserID {
		return nil, nil
	}
	m.souls[s.ID] = s
	return s, nil
}

func (m *mockSoulRepo) Delete(_ context.Context, id, userID int64) error {
	s, ok := m.souls[id]
	if !ok || s.UserID != userID {
		return nil
	}
	delete(m.souls, id)
	return nil
}

// mockSoulHistoryRepo for testing
type mockSoulHistoryRepo struct {
	histories []*model.SoulHistory
	nextID    int64
}

func newMockSoulHistoryRepo() *mockSoulHistoryRepo {
	return &mockSoulHistoryRepo{nextID: 1}
}

func (m *mockSoulHistoryRepo) Create(_ context.Context, h *model.SoulHistory) (*model.SoulHistory, error) {
	h.ID = m.nextID
	m.nextID++
	m.histories = append(m.histories, h)
	return h, nil
}

func (m *mockSoulHistoryRepo) ListBySoulID(_ context.Context, soulID int64, limit int) ([]*model.SoulHistory, error) {
	var result []*model.SoulHistory
	for _, h := range m.histories {
		if h.SoulID == soulID {
			result = append(result, h)
		}
	}
	if len(result) > limit {
		result = result[len(result)-limit:]
	}
	return result, nil
}

func TestSoulCreate_Success(t *testing.T) {
	svc := NewSoulService(newMockSoulRepo(), nil)
	soul, err := svc.Create(context.Background(), 1, "爸爸", "父亲", "温柔", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if soul.Name != "爸爸" {
		t.Errorf("name = %q", soul.Name)
	}
}

func TestSoulCreate_EmptyName(t *testing.T) {
	svc := NewSoulService(newMockSoulRepo(), nil)
	_, err := svc.Create(context.Background(), 1, "", "", "", "", "")
	if err != ErrSoulName {
		t.Fatalf("expected ErrSoulName, got %v", err)
	}
}

func TestSoulCreate_LimitReached(t *testing.T) {
	repo := newMockSoulRepo()
	svc := NewSoulService(repo, nil)
	for i := 0; i < maxSoulsPerUser; i++ {
		svc.Create(context.Background(), 1, "Soul", "", "", "", "")
	}
	_, err := svc.Create(context.Background(), 1, "Extra", "", "", "", "")
	if err != ErrSoulLimit {
		t.Fatalf("expected ErrSoulLimit, got %v", err)
	}
}

func TestSoulGet_Success(t *testing.T) {
	repo := newMockSoulRepo()
	svc := NewSoulService(repo, nil)
	created, _ := svc.Create(context.Background(), 1, "爸爸", "", "", "", "")

	soul, err := svc.Get(context.Background(), 1, created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if soul.Name != "爸爸" {
		t.Errorf("name = %q", soul.Name)
	}
}

func TestSoulGet_NotFound(t *testing.T) {
	svc := NewSoulService(newMockSoulRepo(), nil)
	_, err := svc.Get(context.Background(), 1, 999)
	if err != ErrSoulNotFound {
		t.Fatalf("expected ErrSoulNotFound, got %v", err)
	}
}

func TestSoulGet_Forbidden(t *testing.T) {
	repo := newMockSoulRepo()
	svc := NewSoulService(repo, nil)
	created, _ := svc.Create(context.Background(), 1, "爸爸", "", "", "", "")

	_, err := svc.Get(context.Background(), 999, created.ID)
	if err != ErrSoulForbid {
		t.Fatalf("expected ErrSoulForbid, got %v", err)
	}
}

func TestSoulList_Empty(t *testing.T) {
	svc := NewSoulService(newMockSoulRepo(), nil)
	souls, err := svc.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(souls) != 0 {
		t.Errorf("expected empty list, got %d", len(souls))
	}
}

func TestSoulUpdate_Success(t *testing.T) {
	repo := newMockSoulRepo()
	historyRepo := newMockSoulHistoryRepo()
	svc := NewSoulService(repo, historyRepo)
	created, _ := svc.Create(context.Background(), 1, "爸爸", "", "", "", "")

	updated, err := svc.Update(context.Background(), 1, created.ID, "父亲", "爸爸", "温柔善良", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "父亲" {
		t.Errorf("name = %q", updated.Name)
	}
	// Verify history was recorded
	if len(historyRepo.histories) == 0 {
		t.Error("expected history records for changes")
	}
}

func TestSoulUpdate_Forbidden(t *testing.T) {
	repo := newMockSoulRepo()
	svc := NewSoulService(repo, nil)
	created, _ := svc.Create(context.Background(), 1, "爸爸", "", "", "", "")

	_, err := svc.Update(context.Background(), 999, created.ID, "Changed", "", "", "", "")
	if err != ErrSoulForbid {
		t.Fatalf("expected ErrSoulForbid, got %v", err)
	}
}

func TestSoulDelete_Success(t *testing.T) {
	repo := newMockSoulRepo()
	svc := NewSoulService(repo, nil)
	created, _ := svc.Create(context.Background(), 1, "爸爸", "", "", "", "")

	err := svc.Delete(context.Background(), 1, created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSoulDelete_Forbidden(t *testing.T) {
	repo := newMockSoulRepo()
	svc := NewSoulService(repo, nil)
	created, _ := svc.Create(context.Background(), 1, "爸爸", "", "", "", "")

	err := svc.Delete(context.Background(), 999, created.ID)
	if err != ErrSoulForbid {
		t.Fatalf("expected ErrSoulForbid, got %v", err)
	}
}

func TestSoulHistory(t *testing.T) {
	repo := newMockSoulRepo()
	historyRepo := newMockSoulHistoryRepo()
	svc := NewSoulService(repo, historyRepo)
	created, _ := svc.Create(context.Background(), 1, "爸爸", "", "", "", "")

	// Make some updates to generate history
	svc.Update(context.Background(), 1, created.ID, "父亲", "爸爸", "", "", "")
	svc.Update(context.Background(), 1, created.ID, "父亲", "爸爸", "温柔", "慢慢说", "")

	history, err := svc.History(context.Background(), 1, created.ID, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(history) == 0 {
		t.Error("expected non-empty history")
	}
}
