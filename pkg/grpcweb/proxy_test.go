package grpcweb

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWrapHandler_GRPCWeb(t *testing.T) {
	handler := WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/service.API/Method", nil)
	req.Header.Set("Content-Type", "application/grpc-web")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Type") != "application/grpc-web+proto" {
		t.Errorf("expected application/grpc-web+proto, got %s", rec.Header().Get("Content-Type"))
	}
	if rec.Header().Get("Access-Control-Expose-Headers") == "" {
		t.Error("expected grpc-status in exposed headers")
	}
}
