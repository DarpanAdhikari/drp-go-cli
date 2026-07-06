package routes

import (
	"database/sql"
	"net/http"

	"{{.ModuleName}}/internal/{{.DomainName}}"
)

// Register{{.Name}}Routes wires the {{.Name}} resource routes onto mux.
// Call this from your main router setup.
func Register{{.Name}}Routes(mux *http.ServeMux, db *sql.DB) {
	repo := {{.DomainName}}.New{{.Name}}Repository(db)
	svc := {{.DomainName}}.New{{.Name}}Service(repo)
	h := {{.DomainName}}.New{{.Name}}Handler(svc)

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
