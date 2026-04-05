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

type SoulRepository interface {
	Create(ctx context.Context, s *model.Soul) (*model.Soul, error)
	GetByID(ctx context.Context, id int64) (*model.Soul, error)
	ListByUserID(ctx context.Context, userID int64) ([]*model.Soul, error)
	Update(ctx context.Context, s *model.Soul) (*model.Soul, error)
	Delete(ctx context.Context, id, userID int64) error
}

type SoulService struct {
	repo SoulRepository
}

func NewSoulService(repo SoulRepository) *SoulService {
	return &SoulService{repo: repo}
}

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
