CREATE TABLE IF NOT EXISTS souls (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    relation TEXT NOT NULL DEFAULT '',
    personality TEXT NOT NULL DEFAULT '',
    speech_style TEXT NOT NULL DEFAULT '',
    memory TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_souls_user_id ON souls(user_id);
