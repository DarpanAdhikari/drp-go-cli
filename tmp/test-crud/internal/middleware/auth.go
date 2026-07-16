package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"/tmp/test-crud/internal/auth"
	"/tmp/test-crud/internal/config"
	"/tmp/test-crud/internal/shared"
)

// Auth returns middleware that verifies the Bearer token and enriches the
// request context with user_id and token_jti.
func Auth(cfg *config.Config, db *sql.DB, next http.Handler) http.Handler {
	tokenStore := auth.NewTokenStore(db)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			writeError(w, http.StatusUnauthorized, "authorization required")
			return
		}

		claims, err := auth.ParseToken(parts[1], cfg)
		if err != nil || claims.TokenType != auth.TokenTypeAccess {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		if err := tokenStore.EnsureActive(claims.ID, auth.TokenTypeAccess); err != nil {
			writeError(w, http.StatusUnauthorized, "token revoked or expired")
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, shared.UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, shared.TokenJTIKey, claims.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
