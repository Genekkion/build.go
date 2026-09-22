package db

import (
	"context"
	"database/sql"
)

type ctxKeyDB struct{}

func WithDB(ctx context.Context, db *sql.DB) context.Context {
	return context.WithValue(ctx, ctxKeyDB{}, db)
}

func DBFromCtx(ctx context.Context) *sql.DB {
	if db, ok := ctx.Value(ctxKeyDB{}).(*sql.DB); ok {
		return db
	}
	return nil
}
