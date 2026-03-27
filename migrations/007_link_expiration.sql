-- +goose Up
ALTER TABLE links ADD COLUMN expires_at TIMESTAMPTZ;
ALTER TABLE links ADD COLUMN max_clicks BIGINT;
ALTER TABLE links ADD COLUMN click_count BIGINT NOT NULL DEFAULT 0;

CREATE INDEX idx_links_expires_at ON links (expires_at) WHERE expires_at IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_links_expires_at;
ALTER TABLE links DROP COLUMN IF EXISTS click_count;
ALTER TABLE links DROP COLUMN IF EXISTS max_clicks;
ALTER TABLE links DROP COLUMN IF EXISTS expires_at;
