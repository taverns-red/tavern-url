-- +goose Up
CREATE TABLE applications (
    id             BIGSERIAL    PRIMARY KEY,
    org_name       VARCHAR(255) NOT NULL,
    mission        TEXT         NOT NULL,
    website        VARCHAR(255),
    contact_email  VARCHAR(255) NOT NULL,
    contact_name   VARCHAR(255) NOT NULL,
    status         VARCHAR(20)  NOT NULL DEFAULT 'pending', -- pending, approved, rejected
    reviewed_by    BIGINT       REFERENCES users(id),
    reviewed_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_applications_status ON applications (status);

ALTER TABLE users ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS is_admin;
DROP TABLE IF EXISTS applications;
