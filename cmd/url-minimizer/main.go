package main

import (
	"log/slog"
	"net/http"
	"os"
	"url-minimizer/internal/config"
	"url-minimizer/internal/http-server/handlers/save"
	"url-minimizer/internal/storage/sqlite"
	"url-minimizer/pkg/logger/sl"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.Load()

	log := setLogger(cfg.LogLevel)

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("Failed to initialize storage", sl.Err(err))
	}
	defer func() {
		if err := storage.Close(); err != nil {
			log.Error("Failed to close storage", sl.Err(err))
		}
	}()

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/", save.New(log, storage))

	log.Info("Starting server", slog.String("address", cfg.Address))
	log.Debug("Debug logs enabled")

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HttpServer.Timeout,
		WriteTimeout: cfg.HttpServer.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("Failed to start server", slog.String("address", cfg.Address))
	}

	log.Error("Server stopped")
}

func setLogger(logLevel string) *slog.Logger {
	var log *slog.Logger

	switch logLevel {
	case "debug":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "info":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	}

	return log
}
