package logging_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/logging"
)

func newLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(logging.NewHandler(slog.NewTextHandler(buf, nil)))
}

func TestHandler_AddsRequestID(t *testing.T) {
	var buf bytes.Buffer

	newLogger(&buf).InfoContext(logging.WithRequestID(context.Background(), "abc"), "hello")

	if !strings.Contains(buf.String(), "request_id=abc") {
		t.Errorf("missing request id: %s", buf.String())
	}
}

func TestHandler_NoRequestID(t *testing.T) {
	var buf bytes.Buffer

	newLogger(&buf).InfoContext(context.Background(), "hello")

	if strings.Contains(buf.String(), "request_id") {
		t.Errorf("unexpected request id: %s", buf.String())
	}
}

func TestHandler_KeepsRequestIDAfterWith(t *testing.T) {
	var buf bytes.Buffer

	l := newLogger(&buf).With("a", 1).WithGroup("g")
	l.InfoContext(logging.WithRequestID(context.Background(), "abc"), "hello", "b", 2)

	if !strings.Contains(buf.String(), "request_id=abc") {
		t.Errorf("missing request id: %s", buf.String())
	}
}
