-- +goose Up
CREATE TABLE password_reset_tokens (
    id         BIGSERIAL    PRIMARY KEY,
    user_id    BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64)  NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ  NOT NULL,
    used       BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reset_tokens_hash ON password_reset_tokens (token_hash);
CREATE INDEX idx_reset_tokens_user ON password_reset_tokens (user_id);

-- +goose Down
DROP TABLE IF EXISTS password_reset_tokens;
