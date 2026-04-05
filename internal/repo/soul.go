package repo

import (
	"context"
	"database/sql"

	"github.com/wozai/wozai/internal/model"
)

// SoulRepo implements the SoulRepository interface.
type SoulRepo struct {
	db *sql.DB
}

// NewSoulRepo creates a new SoulRepo.
func NewSoulRepo(db *sql.DB) *SoulRepo {
	return &SoulRepo{db: db}
}

// Create inserts a new soul.
func (r *SoulRepo) Create(ctx context.Context, s *model.Soul) (*model.Soul, error) {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO souls (user_id, name, relation, personality, speech_style, memory)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at, updated_at`,
		s.UserID, s.Name, s.Relation, s.Personality, s.SpeechStyle, s.Memory,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// GetByID retrieves a soul by ID.
func (r *SoulRepo) GetByID(ctx context.Context, id int64) (*model.Soul, error) {
	var s model.Soul
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, relation, personality, speech_style, memory, created_at, updated_at
		 FROM souls WHERE id = $1`,
		id,
	).Scan(&s.ID, &s.UserID, &s.Name, &s.Relation, &s.Personality, &s.SpeechStyle, &s.Memory, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ListByUserID returns all souls for a user.
func (r *SoulRepo) ListByUserID(ctx context.Context, userID int64) ([]*model.Soul, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, name, relation, personality, speech_style, memory, created_at, updated_at
		 FROM souls WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var souls []*model.Soul
	for rows.Next() {
		s := &model.Soul{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Relation, &s.Personality, &s.SpeechStyle, &s.Memory, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		souls = append(souls, s)
	}
	return souls, rows.Err()
}

// Update updates a soul.
func (r *SoulRepo) Update(ctx context.Context, s *model.Soul) (*model.Soul, error) {
	err := r.db.QueryRowContext(ctx,
		`UPDATE souls SET name=$1, relation=$2, personality=$3, speech_style=$4, memory=$5, updated_at=now()
		 WHERE id=$6 AND user_id=$7
		 RETURNING id, user_id, name, relation, personality, speech_style, memory, created_at, updated_at`,
		s.Name, s.Relation, s.Personality, s.SpeechStyle, s.Memory, s.ID, s.UserID,
	).Scan(&s.ID, &s.UserID, &s.Name, &s.Relation, &s.Personality, &s.SpeechStyle, &s.Memory, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Delete removes a soul.
func (r *SoulRepo) Delete(ctx context.Context, id, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM souls WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	return err
}
