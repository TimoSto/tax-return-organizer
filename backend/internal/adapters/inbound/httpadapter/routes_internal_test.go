package httpadapter

import (
	"net/http"
	"strings"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
)

func TestCheckPathTenants(t *testing.T) {
	ok := []httpx.Route{
		{Pattern: "GET /collectors/{collectorId}", Tenant: true},
		{Pattern: "POST /collectors"},
	}
	if err := checkPathTenants(ok); err != nil {
		t.Errorf("err = %v, want nil", err)
	}

	bad := []httpx.Route{
		{Pattern: "GET /collectors/{collectorId}/x"},
		{Pattern: "GET /y", Tenant: true},
	}

	err := checkPathTenants(bad)
	if err == nil || !strings.Contains(err.Error(), "/x") || !strings.Contains(err.Error(), "GET /y") {
		t.Errorf("err = %v, want both routes named", err)
	}
}

func TestWithTenant_SkipsRoutesWithoutTenantFlag(t *testing.T) {
	called := false
	h := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })

	got := withTenant(httpx.Route{Pattern: "GET /x", Handler: h}, nil)
	got.ServeHTTP(nil, nil)

	if !called {
		t.Error("handler not called directly")
	}
}
