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

func TestSoulCreate_Success(t *testing.T) {
	svc := NewSoulService(newMockSoulRepo())
	soul, err := svc.Create(context.Background(), 1, "爸爸", "父亲", "温和", "慢条斯理", "喜欢钓鱼")
	if err != nil {
		t.Fatal(err)
	}
	if soul.ID == 0 || soul.Name != "爸爸" {
		t.Errorf("unexpected soul: %+v", soul)
	}
}

func TestSoulCreate_EmptyName(t *testing.T) {
	svc := NewSoulService(newMockSoulRepo())
	_, err := svc.Create(context.Background(), 1, "", "父亲", "", "", "")
	if err != ErrSoulName {
		t.Errorf("err = %v; want ErrSoulName", err)
	}
}

func TestSoulCreate_LimitReached(t *testing.T) {
	r := newMockSoulRepo()
	svc := NewSoulService(r)
	for i := 0; i < maxSoulsPerUser; i++ {
		r.souls[int64(i+1)] = &model.Soul{ID: int64(i + 1), UserID: 1, Name: "soul"}
		r.nextID = int64(i + 2)
	}
	_, err := svc.Create(context.Background(), 1, "extra", "", "", "", "")
	if err != ErrSoulLimit {
		t.Errorf("err = %v; want ErrSoulLimit", err)
	}
}

func TestSoulGet_Success(t *testing.T) {
	r := newMockSoulRepo()
	svc := NewSoulService(r)
	r.souls[1] = &model.Soul{ID: 1, UserID: 10, Name: "妈妈"}
	soul, err := svc.Get(context.Background(), 10, 1)
	if err != nil || soul.Name != "妈妈" {
		t.Errorf("err=%v soul=%+v", err, soul)
	}
}

func TestSoulGet_NotFound(t *testing.T) {
	svc := NewSoulService(newMockSoulRepo())
	_, err := svc.Get(context.Background(), 1, 999)
	if err != ErrSoulNotFound {
		t.Errorf("err = %v; want ErrSoulNotFound", err)
	}
}

func TestSoulGet_Forbidden(t *testing.T) {
	r := newMockSoulRepo()
	svc := NewSoulService(r)
	r.souls[1] = &model.Soul{ID: 1, UserID: 10, Name: "test"}
	_, err := svc.Get(context.Background(), 99, 1)
	if err != ErrSoulForbid {
		t.Errorf("err = %v; want ErrSoulForbid", err)
	}
}

func TestSoulList_Empty(t *testing.T) {
	svc := NewSoulService(newMockSoulRepo())
	souls, err := svc.List(context.Background(), 1)
	if err != nil || len(souls) != 0 {
		t.Errorf("err=%v len=%d; want 0", err, len(souls))
	}
}

func TestSoulUpdate_Success(t *testing.T) {
	r := newMockSoulRepo()
	svc := NewSoulService(r)
	r.souls[1] = &model.Soul{ID: 1, UserID: 10, Name: "旧名"}
	soul, err := svc.Update(context.Background(), 10, 1, "新名", "朋友", "", "", "")
	if err != nil || soul.Name != "新名" {
		t.Errorf("err=%v soul=%+v", err, soul)
	}
}

func TestSoulUpdate_Forbidden(t *testing.T) {
	r := newMockSoulRepo()
	svc := NewSoulService(r)
	r.souls[1] = &model.Soul{ID: 1, UserID: 10, Name: "test"}
	_, err := svc.Update(context.Background(), 99, 1, "名", "", "", "", "")
	if err != ErrSoulForbid {
		t.Errorf("err = %v; want ErrSoulForbid", err)
	}
}

func TestSoulDelete_Success(t *testing.T) {
	r := newMockSoulRepo()
	svc := NewSoulService(r)
	r.souls[1] = &model.Soul{ID: 1, UserID: 10}
	err := svc.Delete(context.Background(), 10, 1)
	if err != nil {
		t.Errorf("err = %v; want nil", err)
	}
}

func TestSoulDelete_Forbidden(t *testing.T) {
	r := newMockSoulRepo()
	svc := NewSoulService(r)
	r.souls[1] = &model.Soul{ID: 1, UserID: 10}
	err := svc.Delete(context.Background(), 99, 1)
	if err != ErrSoulForbid {
		t.Errorf("err = %v; want ErrSoulForbid", err)
	}
}
