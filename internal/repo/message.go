package repo

import (
	"context"
	"database/sql"

	"github.com/wozai/wozai/internal/model"
)

// MessageRepo implements the MessageRepository interface.
type MessageRepo struct {
	db *sql.DB
}

// NewMessageRepo creates a new MessageRepo.
func NewMessageRepo(db *sql.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

// Create inserts a new message.
func (r *MessageRepo) Create(ctx context.Context, m *model.Message) (*model.Message, error) {
	const q = `INSERT INTO messages (soul_id, role, content, sentiment)
		VALUES ($1, $2, $3, COALESCE(NULLIF($4, ''), NULL))
		RETURNING id, created_at`
	err := r.db.QueryRowContext(ctx, q, m.SoulID, m.Role, m.Content, m.Sentiment).Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// ListBySoulID retrieves messages for a soul.
func (r *MessageRepo) ListBySoulID(ctx context.Context, soulID int64, limit int) ([]*model.Message, error) {
	const q = `SELECT id, soul_id, role, content, COALESCE(sentiment, ''), created_at
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
		if err := rows.Scan(&m.ID, &m.SoulID, &m.Role, &m.Content, &m.Sentiment, &m.CreatedAt); err != nil {
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
