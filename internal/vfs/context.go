package vfs

import "context"

type ctxKeyFS struct{}

func WithFS(ctx context.Context, fs FS) context.Context {
	return context.WithValue(ctx, ctxKeyFS{}, fs)
}

func FSFromCtx(ctx context.Context) FS {
	if fs, ok := ctx.Value(ctxKeyFS{}).(FS); ok {
		return fs
	}
	return NewOSFS()
}
