-- +goose Up
CREATE TABLE click_events (
    id              BIGSERIAL   PRIMARY KEY,
    link_id         BIGINT      NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    clicked_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    country         VARCHAR(10) NOT NULL DEFAULT 'unknown',
    device_category VARCHAR(20) NOT NULL DEFAULT 'unknown',
    referrer_domain VARCHAR(255) NOT NULL DEFAULT 'direct'
);

CREATE INDEX idx_click_events_link_id ON click_events (link_id);
CREATE INDEX idx_click_events_clicked_at ON click_events (clicked_at);
CREATE INDEX idx_click_events_link_date ON click_events (link_id, clicked_at);

-- +goose Down
DROP TABLE IF EXISTS click_events;
