package backend_test

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/TimoSto/tax-return-organizer/backend/internal/backend"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	provider, err := testcontainers.NewDockerProvider()
	if err != nil || provider.Health(ctx) != nil {
		t.Skip("docker unavailable")
	}
	defer provider.Close()

	container, err := tcpostgres.Run(ctx, "postgres:17-alpine",
		tcpostgres.WithDatabase("test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	addr := freeAddr(t)

	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)

	go func() { done <- backend.Run(runCtx, backend.Config{DatabaseURL: dsn, HTTPAddr: addr}) }()

	waitForServer(t, addr)

	resp, err := http.Post("http://"+addr+"/collectors", "application/json",
		strings.NewReader(`{"name":"Acme","email":"a@example.com"}`))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()

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

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	return l.Addr().String()
}

func waitForServer(t *testing.T, addr string) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}

		time.Sleep(50 * time.Millisecond)
	}

	t.Fatal("server did not start")
}
