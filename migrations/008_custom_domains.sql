-- +goose Up
CREATE TABLE custom_domains (
    id          BIGSERIAL    PRIMARY KEY,
    org_id      BIGINT       NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    domain      VARCHAR(255) NOT NULL UNIQUE,
    verified    BOOLEAN      NOT NULL DEFAULT FALSE,
    dns_token   VARCHAR(64)  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_custom_domains_org ON custom_domains (org_id);
CREATE INDEX idx_custom_domains_domain ON custom_domains (domain);

-- +goose Down
DROP TABLE IF EXISTS custom_domains;
