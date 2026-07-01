package models

import "time"

// {{.Name}} represents the {{.TableName}} database record.
type {{.Name}} struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// TODO: add your fields here
}
