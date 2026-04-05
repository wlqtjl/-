-- Soul edit history tracking
CREATE TABLE IF NOT EXISTS soul_history (
    id BIGSERIAL PRIMARY KEY,
    soul_id BIGINT NOT NULL REFERENCES souls(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    field_name TEXT NOT NULL,
    old_value TEXT NOT NULL DEFAULT '',
    new_value TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_soul_history_soul_id ON soul_history(soul_id, created_at DESC);
