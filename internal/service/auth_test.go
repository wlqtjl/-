package service

import (
	"context"
	"testing"

	"github.com/wozai/wozai/internal/model"
)

type mockUserRepo struct {
	users  map[string]*model.User
	nextID int64
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*model.User), nextID: 1}
}

func (m *mockUserRepo) Create(_ context.Context, email, passwordHash, nickname string) (*model.User, error) {
	u := &model.User{ID: m.nextID, Email: email, PasswordHash: passwordHash, Nickname: nickname}
	m.nextID++
	m.users[email] = u
	return u, nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (*model.User, error) {
	return m.users[email], nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id int64) (*model.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

const testSecret = "test-secret-key-that-is-at-least-32-chars!!"

func TestRegister_Success(t *testing.T) {
	svc := NewAuthService(newMockUserRepo(), testSecret, 15, 168)
	user, tokens, err := svc.Register(context.Background(), "test@example.com", "password123", "Tester")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil || tokens == nil {
		t.Fatal("user or tokens is nil")
	}
	if user.Email != "test@example.com" {
		t.Errorf("email = %q", user.Email)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Error("empty tokens")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, testSecret, 15, 168)
	_, _, _ = svc.Register(context.Background(), "test@example.com", "password123", "T1")
	_, _, err := svc.Register(context.Background(), "test@example.com", "password456", "T2")
	if err != ErrEmailTaken {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	svc := NewAuthService(newMockUserRepo(), testSecret, 15, 168)
	_, _, err := svc.Register(context.Background(), "test@example.com", "short", "T")
	if err != ErrWeakPassword {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	svc := NewAuthService(newMockUserRepo(), testSecret, 15, 168)
	_, _, err := svc.Register(context.Background(), "noemail", "password123", "T")
	if err != ErrInvalidEmail {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, testSecret, 15, 168)
	_, _, _ = svc.Register(context.Background(), "test@example.com", "password123", "T")

	user, tokens, err := svc.Login(context.Background(), "test@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil || tokens == nil {
		t.Fatal("user or tokens is nil")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, testSecret, 15, 168)
	_, _, _ = svc.Register(context.Background(), "test@example.com", "password123", "T")

	_, _, err := svc.Login(context.Background(), "test@example.com", "wrongpassword")
	if err != ErrInvalidCreds {
		t.Fatalf("expected ErrInvalidCreds, got %v", err)
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	svc := NewAuthService(newMockUserRepo(), testSecret, 15, 168)
	_, _, err := svc.Login(context.Background(), "nobody@example.com", "password123")
	if err != ErrInvalidCreds {
		t.Fatalf("expected ErrInvalidCreds, got %v", err)
	}
}

func TestRefreshToken_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, testSecret, 15, 168)
	_, tokens, _ := svc.Register(context.Background(), "test@example.com", "password123", "T")

	newTokens, err := svc.RefreshToken(context.Background(), tokens.RefreshToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newTokens.AccessToken == "" || newTokens.RefreshToken == "" {
		t.Error("empty tokens after refresh")
	}
}

func TestRefreshToken_WithAccessToken(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, testSecret, 15, 168)
	_, tokens, _ := svc.Register(context.Background(), "test@example.com", "password123", "T")

	_, err := svc.RefreshToken(context.Background(), tokens.AccessToken)
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateAccessToken_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, testSecret, 15, 168)
	_, tokens, _ := svc.Register(context.Background(), "test@example.com", "password123", "T")

	userID, err := svc.ValidateAccessToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != 1 {
		t.Errorf("userID = %d, want 1", userID)
	}
}

func TestValidateAccessToken_InvalidToken(t *testing.T) {
	svc := NewAuthService(newMockUserRepo(), testSecret, 15, 168)
	_, err := svc.ValidateAccessToken("invalid-token")
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}
