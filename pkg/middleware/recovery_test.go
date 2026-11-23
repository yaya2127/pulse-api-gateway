package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoveryMiddleware(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("critical null pointer simulation")
	})

	recovery := RecoveryMiddleware(nil)(panicHandler)
	req := httptest.NewRequest("GET", "/panic", nil)
	rec := httptest.NewRecorder()

	recovery.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 status code after panic, got %d", rec.Code)
	}
}
