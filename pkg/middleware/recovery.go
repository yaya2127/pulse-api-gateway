package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

func RecoveryMiddleware(logger *log.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := debug.Stack()
					if logger != nil {
						logger.Printf("[PANIC RECOVERED] %v\n%s", rec, string(stack))
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, `{"error":"Internal server error","message":"An unexpected error occurred"}`)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
