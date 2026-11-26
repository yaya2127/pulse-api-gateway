package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkCompressionMiddleware(b *testing.B) {
	cm := NewCompressionMiddleware()
	data := []byte("Benchmark payload repeating chunk for high performance evaluation.")
	handler := cm.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
