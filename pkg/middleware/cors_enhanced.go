package middleware

import (
	"net/http"
	"regexp"
	"strconv"
)

type EnhancedCORSConfig struct {
	AllowedOriginRegex *regexp.Regexp
	AllowCredentials   bool
	MaxAgeSeconds      int
	AllowedMethods     string
	AllowedHeaders     string
}

type EnhancedCORS struct {
	cfg EnhancedCORSConfig
}

func NewEnhancedCORS(cfg EnhancedCORSConfig) *EnhancedCORS {
	if cfg.AllowedMethods == "" {
		cfg.AllowedMethods = "GET, POST, PUT, DELETE, OPTIONS, PATCH"
	}
	if cfg.AllowedHeaders == "" {
		cfg.AllowedHeaders = "Authorization, Content-Type, X-Request-ID, traceparent"
	}
	return &EnhancedCORS{cfg: cfg}
}

func (c *EnhancedCORS) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && (c.cfg.AllowedOriginRegex == nil || c.cfg.AllowedOriginRegex.MatchString(origin)) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", c.cfg.AllowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", c.cfg.AllowedHeaders)
			if c.cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if c.cfg.MaxAgeSeconds > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(c.cfg.MaxAgeSeconds))
			}
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
