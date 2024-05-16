// go run cmd/journal/main.go --config="config/local.yaml"

package main

import (
	"net/http"

	"log/slog"
	"os"

	"github.com/arxonic/journal/internal/config"
	"github.com/arxonic/journal/internal/http-server/handlers/url/disciplines"
	"github.com/arxonic/journal/internal/http-server/middleware/auth"
	"github.com/arxonic/journal/internal/lib/logger/sl"
	"github.com/arxonic/journal/internal/storage/sqlite"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// init config
	cfg := config.MustLoad()

	// init logger
	log := setupLogger(cfg.Env)
	log.Info("starting journal", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	// init storage
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	log.Info("storage init successfully")
	_ = storage

	// init router
	router := chi.NewRouter()

	// middleware
	authMiddleware := auth.New(cfg.Secret, storage)
	router.Use(middleware.RequestID)
	router.Use(authMiddleware.Auth)

	// Handlers
	// router.Get("/url", save.New(log, storage))
	router.Get("/disciplines", disciplines.New(log, storage))

	log.Info("staring server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
		// ReadTimeout:  cfg.HTTPServer.Timeout,
		// WriteTimeout: cfg.HTTPServer.Timeout,
		// IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
