package routes

import (
	"database/sql"
	"net/http"

	"{{.ModuleName}}/internal/handlers"
	"{{.ModuleName}}/internal/repositories"
	"{{.ModuleName}}/internal/services"
)

// Register{{.Name}}Routes wires the {{.Name}} resource routes onto mux.
// Call this from your main router setup.
func Register{{.Name}}Routes(mux *http.ServeMux, db *sql.DB) {
	repo := repositories.New{{.Name}}Repository(db)
	svc := services.New{{.Name}}Service(repo)
	h := handlers.New{{.Name}}Handler(svc)

	mux.HandleFunc("/{{.RouteName}}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.Index(w, r)
		case http.MethodPost:
			h.Store(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/{{.RouteName}}/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.Show(w, r)
		case http.MethodPut, http.MethodPatch:
			h.Update(w, r)
		case http.MethodDelete:
			h.Destroy(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
