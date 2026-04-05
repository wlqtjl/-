package repo

import (
	"context"
	"database/sql"

	"github.com/wozai/wozai/internal/model"
)

type MessageRepo struct {
	db *sql.DB
}

func NewMessageRepo(db *sql.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

func (r *MessageRepo) Create(ctx context.Context, m *model.Message) (*model.Message, error) {
	const q = `INSERT INTO messages (soul_id, role, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	err := r.db.QueryRowContext(ctx, q, m.SoulID, m.Role, m.Content).Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *MessageRepo) ListBySoulID(ctx context.Context, soulID int64, limit int) ([]*model.Message, error) {
	const q = `SELECT id, soul_id, role, content, created_at
		FROM messages WHERE soul_id = $1
		ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, q, soulID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*model.Message
	for rows.Next() {
		m := &model.Message{}
		if err := rows.Scan(&m.ID, &m.SoulID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// reverse to chronological order
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}
