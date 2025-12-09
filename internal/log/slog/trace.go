package slog

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// CtxKey is the key for the trace id in the context.
type CtxKey int

const (
	// ctxTraceID is the key for the trace id in the context.
	ctxTraceID CtxKey = iota
)

// CtxWithTraceID adds a trace id to the context.
func CtxWithTraceID(ctx context.Context) context.Context {
	traceId, err := uuid.NewV7()
	if err != nil {
		GetDefaultLogger().Error("Unable to generate trace id",
			"error", err,
		)
		return ctx
	}
	return context.WithValue(ctx, ctxTraceID, traceId)
}

// traceIdFromCtx returns the trace id from the context.
func traceIdFromCtx(ctx context.Context) (*uuid.UUID, error) {
	raw := ctx.Value(ctxTraceID)
	if raw == nil {
		return nil, fmt.Errorf("trace id not found in context")
	}
	v, ok := raw.(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("trace id is of incorrect type, type: %T", raw)
	}
	return &v, nil
}
