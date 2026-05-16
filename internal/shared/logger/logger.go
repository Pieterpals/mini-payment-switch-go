package logger

import (
	"log/slog"
	"os"
)

// Setup initializes the global slog logger based on the application environment.
// - "development": Uses TextHandler with DEBUG level for human-readable output.
// - All other envs: Uses JSONHandler with INFO level for structured production logging.
func Setup(env string) *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if env == "development" {
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
