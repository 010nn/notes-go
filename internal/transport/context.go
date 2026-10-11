package transport

import "context"

type userCtxKey struct{}

// WithUser кладёт username в контекст.
func WithUser(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, userCtxKey{}, username)
}

// UserFromCtx достаёт username из контекста.
func UserFromCtx(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userCtxKey{}).(string)
	return v, ok && v != ""
}
