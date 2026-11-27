package grpcweb

import (
	"fmt"
	"net/http"
	"strings"
)

func IsGRPCWebRequest(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	return strings.HasPrefix(ct, "application/grpc-web")
}

func WrapHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsGRPCWebRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/grpc-web+proto")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "grpc-status, grpc-message")

		next.ServeHTTP(w, r)
	})
}
