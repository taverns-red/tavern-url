-- +goose Up
CREATE TABLE orgs (
    id         BIGSERIAL    PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    slug       VARCHAR(64)  NOT NULL UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE memberships (
    id      BIGSERIAL PRIMARY KEY,
    user_id BIGINT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id  BIGINT    NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    role    VARCHAR(20) NOT NULL DEFAULT 'member',
    UNIQUE(user_id, org_id)
);

-- Add org_id to links (nullable for now, will be required after backfill).
ALTER TABLE links ADD COLUMN org_id BIGINT REFERENCES orgs(id) ON DELETE CASCADE;
ALTER TABLE links ADD COLUMN created_by BIGINT REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX idx_memberships_user ON memberships (user_id);
CREATE INDEX idx_memberships_org ON memberships (org_id);
CREATE INDEX idx_links_org ON links (org_id);

-- +goose Down
ALTER TABLE links DROP COLUMN IF EXISTS created_by;
ALTER TABLE links DROP COLUMN IF EXISTS org_id;
DROP TABLE IF EXISTS memberships;
DROP TABLE IF EXISTS orgs;
