-- Migration: create_{{.TableName}}_table (up)
CREATE TABLE IF NOT EXISTS {{.TableName}} (
    id         BIGSERIAL    PRIMARY KEY,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
    -- TODO: add your columns here
);
