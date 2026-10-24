package logging

import "context"

type ctxKey struct{}

// WithRequestID returns a context that carries the request id, which NewHandler adds to every log record.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// RequestIDFrom returns the request id stored by WithRequestID.
func RequestIDFrom(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxKey{}).(string)

	return id, ok
}
