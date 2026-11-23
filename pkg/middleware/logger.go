package middleware

import (
	"log"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	bytesWritten int64
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := lw.ResponseWriter.Write(b)
	lw.bytesWritten += int64(n)
	return n, err
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func LoggerMiddleware(logger *log.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lw := newLoggingResponseWriter(w)

			next.ServeHTTP(lw, r)

			duration := time.Since(start)
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = "-"
			}

			if logger != nil {
				logger.Printf("[%s] %s %s %d %s %d bytes (req_id: %s)",
					r.Method, r.URL.Path, r.RemoteAddr, lw.statusCode, duration, lw.bytesWritten, reqID)
			}
		})
	}
}
