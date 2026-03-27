-- +goose Up
ALTER TABLE memberships ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active';

CREATE INDEX idx_memberships_status ON memberships (status);

-- +goose Down
DROP INDEX IF EXISTS idx_memberships_status;
ALTER TABLE memberships DROP COLUMN IF EXISTS status;
