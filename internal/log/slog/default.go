package slog

import (
	"log/slog"
	"os"
)

var (
	// defaultLoggerCleanups are functions to be called when the default logger is cleaned up.
	defaultLoggerCleanups []func()

	// defaultLogger is the "global" logger used when no logger is provided.
	defaultLogger = func() *slog.Logger {
		handlers := []slog.Handler{
			NewHandler(os.Stdout, nil),
		}
		return NewLogger(handlers...)
	}()
)

// CleanupDefaultLogger cleans up the default logger.
func CleanupDefaultLogger() {
	for _, f := range defaultLoggerCleanups {
		f()
	}
}

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
