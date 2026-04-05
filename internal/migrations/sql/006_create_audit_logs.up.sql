-- Audit logs for security tracking
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL DEFAULT 0,
    action TEXT NOT NULL,
    resource TEXT NOT NULL,
    detail TEXT DEFAULT '',
    ip TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id, created_at DESC);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);
