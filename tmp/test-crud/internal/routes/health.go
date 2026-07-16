package routes

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// RegisterHealthRoute adds a GET /healthz endpoint that reports database status.
func RegisterHealthRoute(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})
}
