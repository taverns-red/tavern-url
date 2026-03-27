-- +goose Up
ALTER TABLE links ADD COLUMN password_hash VARCHAR(72);

-- +goose Down
ALTER TABLE links DROP COLUMN IF EXISTS password_hash;
