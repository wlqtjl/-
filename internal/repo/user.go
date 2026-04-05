package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/wozai/wozai/internal/model"
)

// UserRepo implements the UserRepository interface.
type UserRepo struct {
	db *sql.DB
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create inserts a new user.
func (r *UserRepo) Create(ctx context.Context, email, passwordHash, nickname string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash, nickname) VALUES ($1, $2, $3)
		 RETURNING id, email, password_hash, nickname, COALESCE(avatar, ''), COALESCE(bio, ''), created_at, updated_at`,
		email, passwordHash, nickname,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Nickname, &u.Avatar, &u.Bio, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

// GetByEmail retrieves a user by email.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, nickname, COALESCE(avatar, ''), COALESCE(bio, ''), created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Nickname, &u.Avatar, &u.Bio, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

// GetByID retrieves a user by ID.
func (r *UserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, nickname, COALESCE(avatar, ''), COALESCE(bio, ''), created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Nickname, &u.Avatar, &u.Bio, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

// Touch updates the updated_at timestamp.
func (r *UserRepo) Touch(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET updated_at = $1 WHERE id = $2`,
		time.Now(), id,
	)
	return err
}

// UpdateProfile updates user profile fields.
func (r *UserRepo) UpdateProfile(ctx context.Context, userID int64, nickname, avatar, bio string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRowContext(ctx,
		`UPDATE users SET nickname = $1, avatar = $2, bio = $3, updated_at = now()
		 WHERE id = $4
		 RETURNING id, email, password_hash, nickname, COALESCE(avatar, ''), COALESCE(bio, ''), created_at, updated_at`,
		nickname, avatar, bio, userID,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Nickname, &u.Avatar, &u.Bio, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return &u, nil
}

// UpdatePassword updates the user's password hash.
func (r *UserRepo) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = $1, updated_at = now() WHERE id = $2`,
		passwordHash, userID,
	)
	return err
}

// GetStats returns global statistics.
func (r *UserRepo) GetStats(ctx context.Context) (*model.ChatStats, error) {
	var stats model.ChatStats
	err := r.db.QueryRowContext(ctx,
		`SELECT
			(SELECT COUNT(*) FROM users) AS total_users,
			(SELECT COUNT(*) FROM souls) AS total_souls,
			(SELECT COUNT(*) FROM messages) AS total_messages,
			(SELECT COUNT(DISTINCT user_id) FROM souls WHERE updated_at > now() - interval '24 hours') AS active_today`,
	).Scan(&stats.TotalUsers, &stats.TotalSouls, &stats.TotalMessages, &stats.ActiveToday)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	return &stats, nil
}

// GetUserStats returns stats for a specific user.
func (r *UserRepo) GetUserStats(ctx context.Context, userID int64) (soulsCount, messagesCount int64, err error) {
	err = r.db.QueryRowContext(ctx,
		`SELECT
			(SELECT COUNT(*) FROM souls WHERE user_id = $1),
			(SELECT COUNT(*) FROM messages m JOIN souls s ON m.soul_id = s.id WHERE s.user_id = $1)`,
		userID,
	).Scan(&soulsCount, &messagesCount)
	return
}
