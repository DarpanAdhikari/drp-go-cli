-- Migration: create_{{.TableName}}_table (down)
DROP TABLE IF EXISTS {{.TableName}};
