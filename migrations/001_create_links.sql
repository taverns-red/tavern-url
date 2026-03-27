-- +goose Up
CREATE TABLE links (
    id           BIGSERIAL    PRIMARY KEY,
    slug         VARCHAR(64)  NOT NULL UNIQUE,
    original_url TEXT         NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_links_slug ON links (slug);

-- +goose Down
DROP TABLE IF EXISTS links;
