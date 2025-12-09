package slog

import (
	"context"
	"io"
	"log/slog"
)

// Handler enables the writing of logs to multiple pipes.
type Handler struct {
	subHandlers []slog.Handler
}

// NewLogger creates a new logger.
// Panics if no handlers are provided.
func NewLogger(handlers ...slog.Handler) *slog.Logger {
	if len(handlers) == 0 {
		panic("logger needs at least 1 handler")
	}

	return slog.New(&Handler{
		subHandlers: handlers,
	})
}

// NewHandler creates a new handler.
func NewHandler(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	return slog.NewJSONHandler(w, opts)
}

// AddHandler adds a new handler.
func (h *Handler) AddHandler(handler slog.Handler) {
	h.subHandlers = append(h.subHandlers, handler)
}

// Handle handles a log record.
func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	for _, sh := range h.subHandlers {
		traceId, err := traceIdFromCtx(ctx)
		if err == nil {
			r.Add("traceId", traceId.String())
		}
		err = sh.Handle(ctx, r)
		if err != nil {
			return err
		}
	}

	return nil
}

// Enabled checks if the handler is enabled.
func (h *Handler) Enabled(ctx context.Context, l slog.Level) bool {
	for _, sh := range h.subHandlers {
		if sh.Enabled(ctx, l) {
			return true
		}
	}

	return false
}

// WithAttrs creates a new handler with the given attributes.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.subHandlers))

	for i, sh := range h.subHandlers {
		newHandlers[i] = sh.WithAttrs(attrs)
	}

	return &Handler{
		subHandlers: newHandlers,
	}
}

// WithGroup creates a new handler with the given group name.
func (h *Handler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.subHandlers))

	for i, sh := range h.subHandlers {
		newHandlers[i] = sh.WithGroup(name)
	}

	return &Handler{
		subHandlers: newHandlers,
	}
}
