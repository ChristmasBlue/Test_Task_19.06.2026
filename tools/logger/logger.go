package logger

import (
	"log/slog"
	"os"
	"test_task/tools/config"
)

func InitLogger(cfg config.Config) {
	var logLevel slog.Level

	switch cfg.LoggerLevel() {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: cfg.AddSource(),
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)

	logger := slog.New(handler)

	slog.SetDefault(logger)
}
