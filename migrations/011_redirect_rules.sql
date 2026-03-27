-- +goose Up
CREATE TABLE redirect_rules (
    id              BIGSERIAL    PRIMARY KEY,
    link_id         BIGINT       NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    condition_type  VARCHAR(20)  NOT NULL, -- 'geo_country', 'device_type', 'weighted'
    condition_value VARCHAR(100) NOT NULL, -- e.g., 'US', 'mobile', '50'
    target_url      TEXT         NOT NULL,
    weight          INT          NOT NULL DEFAULT 100,
    priority        INT          NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_redirect_rules_link ON redirect_rules (link_id);

-- +goose Down
DROP TABLE IF EXISTS redirect_rules;
