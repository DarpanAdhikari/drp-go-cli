package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"/tmp/test-crud/internal/config"
	_ "/tmp/test-crud/docs"
	"/tmp/test-crud/internal/middleware"
	"/tmp/test-crud/internal/routes"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Warn(".env not found", "error", err)
	}

	cfg := config.Load()

	var logLevel slog.Level
	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))

	db, err := sql.Open(cfg.DBDriver, cfg.DSN())
	if err != nil {
		slog.Error("db connect", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error("db ping", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to database")

	mux := http.NewServeMux()
	routes.RegisterAuthRoutes(mux, db, cfg)
	routes.RegisterHealthRoute(mux, db)

	// Swagger UI
	mux.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("swagger/doc.json"),
	))

	handler := middleware.RequestID(middleware.CORS(cfg, mux))

	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: handler,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("listening", "port", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
