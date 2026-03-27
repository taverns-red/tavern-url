-- +goose Up
CREATE TABLE webhooks (
    id         BIGSERIAL    PRIMARY KEY,
    org_id     BIGINT       NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    url        TEXT         NOT NULL,
    events     TEXT[]       NOT NULL, -- e.g., '{link.created,link.clicked}'
    secret     VARCHAR(64)  NOT NULL,
    active     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhooks_org ON webhooks (org_id);

-- +goose Down
DROP TABLE IF EXISTS webhooks;
