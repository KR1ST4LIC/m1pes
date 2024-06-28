package app

import (
	"log/slog"
	"m1pes/internal/logging"
	"os"
)

func (a *App) InitLogging() error {
	handler := slog.Handler(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))
	handler = logging.NewSlogWrapper(handler)

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return nil
}
