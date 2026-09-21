package slog

import (
	"log/slog"
	"os"
)

// defaultLogger is the global logger used when no logger is provided.
var defaultLogger = func() *slog.Logger {
	handlers := []slog.Handler{
		NewHandler(os.Stdout, nil),
	}
	return NewLogger(handlers...)
}()

// SetDefaultLogger sets the default logger.
func SetDefaultLogger(logger *slog.Logger) {
	if logger == nil {
		panic("Default logger cannot be nil")
	}
	defaultLogger = logger
}

// GetDefaultLogger returns the default logger.
func GetDefaultLogger() *slog.Logger {
	return defaultLogger
}
