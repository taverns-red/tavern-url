-- +goose Up
CREATE TABLE api_keys (
    id          BIGSERIAL    PRIMARY KEY,
    user_id     BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id      BIGINT       NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    key_hash    VARCHAR(64)  NOT NULL UNIQUE,
    key_prefix  VARCHAR(10)  NOT NULL,
    last_used_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_keys_user ON api_keys (user_id);
CREATE INDEX idx_api_keys_org ON api_keys (org_id);
CREATE INDEX idx_api_keys_hash ON api_keys (key_hash);

-- +goose Down
DROP TABLE IF EXISTS api_keys;
