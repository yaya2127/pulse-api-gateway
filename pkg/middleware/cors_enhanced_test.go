package middleware

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestEnhancedCORS(t *testing.T) {
	cors := NewEnhancedCORS(EnhancedCORSConfig{
		AllowedOriginRegex: regexp.MustCompile(`^https://.*\.pulse\.dev$`),
		MaxAgeSeconds:      86400,
		AllowCredentials:   true,
	})

	handler := cors.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/api/v1/test", nil)
	req.Header.Set("Origin", "https://app.pulse.dev")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("got status %d, want 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Max-Age") != "86400" {
		t.Errorf("got max-age %s, want 86400", rec.Header().Get("Access-Control-Max-Age"))
	}
}
