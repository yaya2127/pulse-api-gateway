package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token format")
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidSig   = errors.New("signature mismatch")
)

type JWTClaims struct {
	Subject   string         `json:"sub"`
	Issuer    string         `json:"iss"`
	Audience  string         `json:"aud"`
	ExpiresAt int64          `json:"exp"`
	IssuedAt  int64          `json:"iat"`
	Role      string         `json:"role"`
	Extra     map[string]any `json:"-"`
}

type JWTValidator struct {
	secret []byte
}

func NewJWTValidator(secret string) *JWTValidator {
	return &JWTValidator{secret: []byte(secret)}
}

func base64URLDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

func base64URLEncode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func (v *JWTValidator) verifySignature(headerPayload, sig string) error {
	mac := hmac.New(sha256.New, v.secret)
	mac.Write([]byte(headerPayload))
	expectedSig := mac.Sum(nil)

	actualSig, err := base64URLDecode(sig)
	if err != nil {
		return ErrInvalidToken
	}

	if !hmac.Equal(expectedSig, actualSig) {
		return ErrInvalidSig
	}
	return nil
}

func (v *JWTValidator) Validate(tokenString string) (*JWTClaims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerPayload := parts[0] + "." + parts[1]
	if err := v.verifySignature(headerPayload, parts[2]); err != nil {
		return nil, err
	}

	payloadBytes, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims JWTClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if claims.ExpiresAt > 0 && time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}

import (
	"context"
)

const ClaimsKey contextKey = "jwt_claims"

func (v *JWTValidator) Middleware(skipPaths []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, p := range skipPaths {
				if strings.HasPrefix(r.URL.Path, p) {
					next.ServeHTTP(w, r)
					return
				}
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintf(w, `{"error":"Unauthorized","message":"Missing or invalid bearer token"}`)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := v.Validate(tokenStr)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintf(w, `{"error":"Unauthorized","message":"%s"}`, err.Error())
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			w.Header().Set("X-User-ID", claims.Subject)
			w.Header().Set("X-User-Role", claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
