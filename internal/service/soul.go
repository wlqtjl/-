package service

import (
	"context"
	"errors"

	"github.com/wozai/wozai/internal/model"
)

var (
	ErrSoulNotFound = errors.New("soul not found")
	ErrSoulForbid   = errors.New("no permission to access this soul")
	ErrSoulLimit    = errors.New("soul limit reached")
	ErrSoulName     = errors.New("soul name is required")
)

const maxSoulsPerUser = 20

// SoulRepository defines data access methods for souls.
type SoulRepository interface {
	Create(ctx context.Context, s *model.Soul) (*model.Soul, error)
	GetByID(ctx context.Context, id int64) (*model.Soul, error)
	ListByUserID(ctx context.Context, userID int64) ([]*model.Soul, error)
	Update(ctx context.Context, s *model.Soul) (*model.Soul, error)
	Delete(ctx context.Context, id, userID int64) error
}

// SoulHistoryRepository defines data access methods for soul edit history.
type SoulHistoryRepository interface {
	Create(ctx context.Context, h *model.SoulHistory) (*model.SoulHistory, error)
	ListBySoulID(ctx context.Context, soulID int64, limit int) ([]*model.SoulHistory, error)
}

// SoulService handles soul business logic.
type SoulService struct {
	repo        SoulRepository
	historyRepo SoulHistoryRepository
}

// NewSoulService creates a new SoulService.
func NewSoulService(repo SoulRepository, historyRepo SoulHistoryRepository) *SoulService {
	return &SoulService{repo: repo, historyRepo: historyRepo}
}

// Create creates a new soul for the given user.
func (s *SoulService) Create(ctx context.Context, userID int64, name, relation, personality, speechStyle, memory string) (*model.Soul, error) {
	if name == "" {
		return nil, ErrSoulName
	}

	existing, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(existing) >= maxSoulsPerUser {
		return nil, ErrSoulLimit
	}

	soul := &model.Soul{
		UserID:      userID,
		Name:        name,
		Relation:    relation,
		Personality: personality,
		SpeechStyle: speechStyle,
		Memory:      memory,
	}
	return s.repo.Create(ctx, soul)
}

// Get retrieves a soul with ownership check.
func (s *SoulService) Get(ctx context.Context, userID, soulID int64) (*model.Soul, error) {
	soul, err := s.repo.GetByID(ctx, soulID)
	if err != nil {
		return nil, err
	}
	if soul == nil {
		return nil, ErrSoulNotFound
	}
	if soul.UserID != userID {
		return nil, ErrSoulForbid
	}
	return soul, nil
}

// List returns all souls for a user.
func (s *SoulService) List(ctx context.Context, userID int64) ([]*model.Soul, error) {
	souls, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if souls == nil {
		souls = []*model.Soul{}
	}
	return souls, nil
}

// Update updates a soul with ownership check and records edit history.
func (s *SoulService) Update(ctx context.Context, userID, soulID int64, name, relation, personality, speechStyle, memory string) (*model.Soul, error) {
	if name == "" {
		return nil, ErrSoulName
	}

	soul, err := s.repo.GetByID(ctx, soulID)
	if err != nil {
		return nil, err
	}
	if soul == nil {
		return nil, ErrSoulNotFound
	}
	if soul.UserID != userID {
		return nil, ErrSoulForbid
	}

	// Record edit history for changed fields
	if s.historyRepo != nil {
		changes := map[string][2]string{
			"name":         {soul.Name, name},
			"relation":     {soul.Relation, relation},
			"personality":  {soul.Personality, personality},
			"speech_style": {soul.SpeechStyle, speechStyle},
			"memory":       {soul.Memory, memory},
		}
		for field, vals := range changes {
			if vals[0] != vals[1] {
				s.historyRepo.Create(ctx, &model.SoulHistory{
					SoulID:    soulID,
					UserID:    userID,
					FieldName: field,
					OldValue:  vals[0],
					NewValue:  vals[1],
				})
			}
		}
	}

	soul.Name = name
	soul.Relation = relation
	soul.Personality = personality
	soul.SpeechStyle = speechStyle
	soul.Memory = memory

	updated, err := s.repo.Update(ctx, soul)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, ErrSoulNotFound
	}
	return updated, nil
}

// Delete deletes a soul with ownership check.
func (s *SoulService) Delete(ctx context.Context, userID, soulID int64) error {
	soul, err := s.repo.GetByID(ctx, soulID)
	if err != nil {
		return err
	}
	if soul == nil {
		return ErrSoulNotFound
	}
	if soul.UserID != userID {
		return ErrSoulForbid
	}
	return s.repo.Delete(ctx, soulID, userID)
}

// History returns edit history for a soul.
func (s *SoulService) History(ctx context.Context, userID, soulID int64, limit int) ([]*model.SoulHistory, error) {
	soul, err := s.repo.GetByID(ctx, soulID)
	if err != nil {
		return nil, err
	}
	if soul == nil {
		return nil, ErrSoulNotFound
	}
	if soul.UserID != userID {
		return nil, ErrSoulForbid
	}
	if s.historyRepo == nil {
		return []*model.SoulHistory{}, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.historyRepo.ListBySoulID(ctx, soulID, limit)
}
