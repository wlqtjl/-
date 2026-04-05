package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/wozai/wozai/internal/model"
)

// SoulHistoryRepo implements the SoulHistoryRepository interface.
type SoulHistoryRepo struct {
	db *sql.DB
}

// NewSoulHistoryRepo creates a new SoulHistoryRepo.
func NewSoulHistoryRepo(db *sql.DB) *SoulHistoryRepo {
	return &SoulHistoryRepo{db: db}
}

// Create inserts a new soul history record.
func (r *SoulHistoryRepo) Create(ctx context.Context, h *model.SoulHistory) (*model.SoulHistory, error) {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO soul_history (soul_id, user_id, field_name, old_value, new_value)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		h.SoulID, h.UserID, h.FieldName, h.OldValue, h.NewValue,
	).Scan(&h.ID, &h.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create soul history: %w", err)
	}
	return h, nil
}

// ListBySoulID returns edit history for a soul.
func (r *SoulHistoryRepo) ListBySoulID(ctx context.Context, soulID int64, limit int) ([]*model.SoulHistory, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, soul_id, user_id, field_name, old_value, new_value, created_at
		 FROM soul_history WHERE soul_id = $1
		 ORDER BY created_at DESC LIMIT $2`,
		soulID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []*model.SoulHistory
	for rows.Next() {
		h := &model.SoulHistory{}
		if err := rows.Scan(&h.ID, &h.SoulID, &h.UserID, &h.FieldName, &h.OldValue, &h.NewValue, &h.CreatedAt); err != nil {
			return nil, err
		}
		histories = append(histories, h)
	}
	return histories, rows.Err()
}

// AuditLogRepo implements audit log storage.
type AuditLogRepo struct {
	db *sql.DB
}

// NewAuditLogRepo creates a new AuditLogRepo.
func NewAuditLogRepo(db *sql.DB) *AuditLogRepo {
	return &AuditLogRepo{db: db}
}

// Create inserts an audit log entry.
func (r *AuditLogRepo) Create(ctx context.Context, log *model.AuditLog) (*model.AuditLog, error) {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO audit_logs (user_id, action, resource, detail, ip)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		log.UserID, log.Action, log.Resource, log.Detail, log.IP,
	).Scan(&log.ID, &log.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create audit log: %w", err)
	}
	return log, nil
}

// List retrieves audit logs with pagination.
func (r *AuditLogRepo) List(ctx context.Context, limit, offset int) ([]*model.AuditLog, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, action, resource, detail, ip, created_at
		 FROM audit_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*model.AuditLog
	for rows.Next() {
		l := &model.AuditLog{}
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.Resource, &l.Detail, &l.IP, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

// ListByUserID retrieves audit logs for a specific user.
func (r *AuditLogRepo) ListByUserID(ctx context.Context, userID int64, limit int) ([]*model.AuditLog, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, action, resource, detail, ip, created_at
		 FROM audit_logs WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*model.AuditLog
	for rows.Next() {
		l := &model.AuditLog{}
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.Resource, &l.Detail, &l.IP, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
