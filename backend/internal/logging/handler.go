package logging

import (
	"context"
	"log/slog"
)

// RequestIDKey is the attribute under which the request id is logged.
const RequestIDKey = "request_id"

// NewHandler wraps h so that every record logged with a context carrying a request id gets a request_id attribute.
func NewHandler(h slog.Handler) slog.Handler { return ctxHandler{h} }

type ctxHandler struct{ slog.Handler }

func (h ctxHandler) Handle(ctx context.Context, r slog.Record) error {
	if id, ok := RequestIDFrom(ctx); ok {
		r.AddAttrs(slog.String(RequestIDKey, id))
	}

	return h.Handler.Handle(ctx, r) //nolint:wrapcheck // a wrapping handler passes the error through unchanged
}

func (h ctxHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return ctxHandler{h.Handler.WithAttrs(attrs)}
}

func (h ctxHandler) WithGroup(name string) slog.Handler {
	return ctxHandler{h.Handler.WithGroup(name)}
}
