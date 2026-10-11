package backend_test

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/TimoSto/tax-return-organizer/backend/internal/backend"
	"github.com/TimoSto/tax-return-organizer/backend/internal/testutil/pgtest"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	if !pgtest.DockerAvailable(ctx) {
		t.Skip("docker unavailable")
	}

	dsn, terminate, err := pgtest.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(terminate)

	addr := freeAddr(t)

	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)

	go func() { done <- backend.Run(runCtx, backend.Config{DatabaseURL: dsn, HTTPAddr: addr}) }()

	waitForServer(t, addr)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+addr+"/collectors",
		strings.NewReader(`{"name":"Acme","email":"a@example.com"}`))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}

	if resp.Header.Get("Location") == "" {
		t.Error("missing Location header")
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("Run did not stop after cancel")
	}
}

func freeAddr(t *testing.T) string {
	t.Helper()

	l, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()

	return l.Addr().String()
}

func waitForServer(t *testing.T, addr string) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := (&net.Dialer{Timeout: 100 * time.Millisecond}).DialContext(context.Background(), "tcp", addr)
		if err == nil {
			_ = conn.Close()
			return
		}

		time.Sleep(50 * time.Millisecond)
	}

	t.Fatal("server did not start")
}
