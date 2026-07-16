package routes

import (
	"database/sql"
	"net/http"

	"golang.org/x/time/rate"

	"/tmp/test-crud/internal/auth"
	"/tmp/test-crud/internal/config"
	"/tmp/test-crud/internal/middleware"
	"/tmp/test-crud/internal/user"
)

// RegisterAuthRoutes wires all auth and user routes onto the mux.
func RegisterAuthRoutes(mux *http.ServeMux, db *sql.DB, cfg *config.Config) {
	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo)
	authHandler := auth.NewHandler(db, userSvc, cfg)
	userHandler := user.NewHandler(userSvc)

	// Rate limiting for auth endpoints
	var authRateLimit func(http.Handler) http.Handler
	if cfg.RateLimitEnabled {
		rl := middleware.NewIPRateLimiter(rate.Limit(cfg.RateLimitRPS), cfg.RateLimitBurst)
		authRateLimit = middleware.RateLimit(rl)
	} else {
		authRateLimit = func(next http.Handler) http.Handler { return next }
	}

	// Auth routes (with rate limiting on mutation endpoints)
	mux.Handle("POST /auth/register", authRateLimit(http.HandlerFunc(authHandler.Register)))
	mux.Handle("POST /auth/login", authRateLimit(http.HandlerFunc(authHandler.Login)))
	mux.Handle("POST /auth/refresh", authRateLimit(http.HandlerFunc(authHandler.Refresh)))
	mux.Handle("POST /auth/logout", middleware.Auth(cfg, db, http.HandlerFunc(authHandler.Logout)))

	// User routes
	mux.Handle("GET /me", middleware.Auth(cfg, db, http.HandlerFunc(userHandler.Me)))
}
