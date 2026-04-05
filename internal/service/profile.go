package service

import (
	"context"
	"database/sql"

	"github.com/wozai/wozai/internal/model"
)

// ProfileRepository defines data access for user profiles.
type ProfileRepository interface {
	GetByID(ctx context.Context, id int64) (*model.User, error)
	UpdateProfile(ctx context.Context, userID int64, nickname, avatar, bio string) (*model.User, error)
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error
}

// StatsRepository defines data access for statistics.
type StatsRepository interface {
	GetStats(ctx context.Context) (*model.ChatStats, error)
	GetUserStats(ctx context.Context, userID int64) (soulsCount, messagesCount int64, err error)
}

// ProfileService handles user profile operations.
type ProfileService struct {
	profileRepo ProfileRepository
}

// NewProfileService creates a new ProfileService.
func NewProfileService(profileRepo ProfileRepository) *ProfileService {
	return &ProfileService{profileRepo: profileRepo}
}

// GetProfile returns a user's profile information.
func (s *ProfileService) GetProfile(ctx context.Context, userID int64) (*model.User, error) {
	user, err := s.profileRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidToken
	}
	return user, nil
}

// UpdateProfile updates a user's profile information.
func (s *ProfileService) UpdateProfile(ctx context.Context, userID int64, nickname, avatar, bio string) (*model.User, error) {
	if s.profileRepo == nil {
		return nil, sql.ErrNoRows
	}
	return s.profileRepo.UpdateProfile(ctx, userID, nickname, avatar, bio)
}
