// go run cmd/journal/main.go --config="config/local.yaml"

package main

import (
	"net/http"

	"log/slog"
	"os"

	"github.com/arxonic/journal/internal/config"
	"github.com/arxonic/journal/internal/http-server/handlers/url/courses"
	"github.com/arxonic/journal/internal/http-server/middleware/auth"
	"github.com/arxonic/journal/internal/lib/logger/sl"
	"github.com/arxonic/journal/internal/services/policy"
	"github.com/arxonic/journal/internal/storage/sqlite"
	"github.com/go-chi/chi/v5"
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

	// Init access list
	accessControl := policy.New()

	// Init router
	router := chi.NewRouter()

	// Middleware
	authMiddleware := auth.New(cfg.Secret, storage)
	router.Use(authMiddleware.Auth)

	// Handlers
	url := "/courses/create"
	accessControl.Add(url, "admin")
	router.Post(url, courses.Create(url, log, storage, accessControl))

	url = "/courses/{courseID}/modify/students"
	accessControl.Add(url, "admin")
	router.Post(url, courses.EnrollStudents(url, log, storage, accessControl))
	router.Delete(url, courses.RemoveStudents(url, log, storage, accessControl))

	// Start server
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
