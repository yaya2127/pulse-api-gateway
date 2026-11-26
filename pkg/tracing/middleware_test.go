package tracing

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTracingMiddleware(t *testing.T) {
	mw := NewMiddleware(nil)
	var captured *TraceContext

	handler := mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/resource", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if captured == nil {
		t.Fatal("expected TraceContext in request context")
	}

	respHeader := rec.Header().Get(HeaderTraceParent)
	if respHeader == "" {
		t.Fatal("expected traceparent in response header")
	}
}
