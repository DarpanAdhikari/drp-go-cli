-- Migration: create_{{.TableName}}_table (up)
CREATE TABLE IF NOT EXISTS {{.TableName}} (
    id         {{if eq .DBDriver "mysql"}}BIGINT UNSIGNED AUTO_INCREMENT{{else}}BIGSERIAL{{end}} PRIMARY KEY,
    created_at {{if eq .DBDriver "mysql"}}DATETIME(6){{else}}TIMESTAMPTZ{{end}} NOT NULL DEFAULT {{if eq .DBDriver "mysql"}}CURRENT_TIMESTAMP(6){{else}}NOW(){{end}},
    updated_at {{if eq .DBDriver "mysql"}}DATETIME(6){{else}}TIMESTAMPTZ{{end}} NOT NULL DEFAULT {{if eq .DBDriver "mysql"}}CURRENT_TIMESTAMP(6){{else}}NOW(){{end}}
    -- TODO: add your columns here
);
