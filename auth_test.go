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
	user, tokens, err := svc.Register(context.Background(), "test@example.com", "password123", "TestUser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("email = %q; want %q", user.Email, "test@example.com")
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Error("expected non-empty tokens")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := NewAuthService(newMockUserRepo(), testSecret, 15, 168)
	_, _, _ = svc.Register(context.Background(), "test@example.com", "password123", "User1")
	_, _, err := svc.Register(context.Background(), "test@example.com", "password456", "User2")
	if err != ErrEmailTaken {
		t.Errorf("err = %v; want ErrEmailTaken", err)
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	svc := NewAuthService(newMockUserRepo(), testSecret, 15, 168)
	_, _, err := svc.Register(context.Background(), "test@example.com", "short", "User")
	if err != ErrWeakPassword {
		t.Errorf("err = %v; want ErrWeakPassword", err)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	svc := NewAuthService(newMockUserRepo(), testSecret, 15, 168)
	_, _, err := svc.Register(context.Background(), "bademail", "password123", "User")
	if err != ErrInvalidEmail {
		t.Errorf("err = %v; want ErrInvalidEmail", err)
	}
}

func TestLogin_Success(t *testing.T) {
	svc := NewAuthService(newMockUserRepo(), testSecret, 15, 168)
	_, _, _ = svc.Register(context.Background(), "test@example.com", "password123", "User")
	user, tokens, err := svc.Login(context.Background(), "test@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("email mismatch")
	}
	if tokens.AccessToken == "" {
		t.Error("expected access token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := NewAuthService(newMockUserRepo(), testSecret, 15, 168)
	_, _, _ = svc.Register(context.Background(), "test@example.com", "password123", "User")
	_, _, err := svc.Login(context.Background(), "test@example.com", "wrongpassword")
	if err != ErrInvalidCreds {
		t.Errorf("err = %v; want ErrInvalidCreds", err)
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	svc := NewAuthService(newMockUserRepo(), testSecret, 15, 168)
	_, _, err := svc.Login(context.Background(), "nobody@example.com", "password123")
	if err != ErrInvalidCreds {
		t.Errorf("err = %v; want ErrInvalidCreds", err)
	}
}

func TestRefreshToken_Success(t *testing.T) {
	svc := NewAuthService(newMockUserRepo(), testSecret, 15, 168)
	_, tokens, _ := svc.Register(context.Background(), "test@example.com", "password123", "User")
	newTokens, err := svc.RefreshToken(context.Background(), tokens.RefreshToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newTokens.AccessToken == "" {
		t.Error("expected new access token")
	}
}

func TestRefreshToken_WithAccessToken(t *testing.T) {
	svc := NewAuthService(newMockUserRepo(), testSecret, 15, 168)
	_, tokens, _ := svc.Register(context.Background(), "test@example.com", "password123", "User")
	_, err := svc.RefreshToken(context.Background(), tokens.AccessToken)
	if err != ErrInvalidToken {
		t.Errorf("err = %v; want ErrInvalidToken (should reject access token as refresh)", err)
	}
}

func TestValidateAccessToken_Success(t *testing.T) {
	svc := NewAuthService(newMockUserRepo(), testSecret, 15, 168)
	_, tokens, _ := svc.Register(context.Background(), "test@example.com", "password123", "User")
	userID, err := svc.ValidateAccessToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != 1 {
		t.Errorf("userID = %d; want 1", userID)
	}
}

func TestValidateAccessToken_InvalidToken(t *testing.T) {
	svc := NewAuthService(newMockUserRepo(), testSecret, 15, 168)
	_, err := svc.ValidateAccessToken("invalid.token.here")
	if err != ErrInvalidToken {
		t.Errorf("err = %v; want ErrInvalidToken", err)
	}
}
