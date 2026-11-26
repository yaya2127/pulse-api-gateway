package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIPFilter_Allowlist(t *testing.T) {
	filter, err := NewIPFilter(ModeAllowlist, []string{"192.168.1.0/24", "10.0.0.0/8"})
	if err != nil {
		t.Fatalf("failed to init filter: %v", err)
	}

	handler := filter.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	reqAllowed := httptest.NewRequest("GET", "/", nil)
	reqAllowed.RemoteAddr = "192.168.1.55:12345"
	recAllowed := httptest.NewRecorder()
	handler.ServeHTTP(recAllowed, reqAllowed)
	if recAllowed.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", recAllowed.Code)
	}

	reqBlocked := httptest.NewRequest("GET", "/", nil)
	reqBlocked.RemoteAddr = "172.16.0.1:54321"
	recBlocked := httptest.NewRecorder()
	handler.ServeHTTP(recBlocked, reqBlocked)
	if recBlocked.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", recBlocked.Code)
	}
}
