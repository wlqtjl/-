package middleware

import (
	"context"

	"github.com/wozai/wozai/internal/model"
	"github.com/wozai/wozai/internal/repo"
)

// AuditLogger records user actions for audit trail.
type AuditLogger struct {
	repo *repo.AuditLogRepo
}

// NewAuditLogger creates a new AuditLogger.
func NewAuditLogger(r *repo.AuditLogRepo) *AuditLogger {
	return &AuditLogger{repo: r}
}

// Log records an audit event.
func (a *AuditLogger) Log(ctx context.Context, userID int64, action, resource, detail, ip string) {
	if a.repo == nil {
		return
	}
	a.repo.Create(ctx, &model.AuditLog{
		UserID:   userID,
		Action:   action,
		Resource: resource,
		Detail:   detail,
		IP:       ip,
	})
}
