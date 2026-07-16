-- Seeder: {{.TableName}}
-- TODO: add your INSERT statements here

-- INSERT INTO {{.TableName}} (created_at, updated_at) VALUES ({{if eq .DBDriver "mysql"}}CURRENT_TIMESTAMP(6){{else}}NOW(){{end}}, {{if eq .DBDriver "mysql"}}CURRENT_TIMESTAMP(6){{else}}NOW(){{end}});
