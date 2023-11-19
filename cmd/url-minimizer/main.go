package main

import (
	"log/slog"
	"os"
	"url-minimizer/internal/config"
)

func main() {
	cfg := config.Load()

	log := setLogger(cfg.LogLevel)

	log.Info("Starting server at", slog.String("address", cfg.Address))
	log.Debug("Debug logs enabled")
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
