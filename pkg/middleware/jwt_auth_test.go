package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

func makeTestToken(secret string, expOffset time.Duration) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	claims := map[string]any{
		"sub": "test_user_42",
		"exp": time.Now().Add(expOffset).Unix(),
		"role": "admin",
	}
	payloadBytes, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	hp := header + "." + payload
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(hp))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return hp + "." + sig
}

func TestJWTValidator(t *testing.T) {
	secret := "test-secret-pulse-2026"
	validator := NewJWTValidator(secret)

	validToken := makeTestToken(secret, time.Hour)
	claims, err := validator.Validate(validToken)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}
	if claims.Subject != "test_user_42" || claims.Role != "admin" {
		t.Errorf("claims mismatch: %+v", claims)
	}

	expiredToken := makeTestToken(secret, -time.Hour)
	_, err = validator.Validate(expiredToken)
	if err != ErrExpiredToken {
		t.Fatalf("expected ErrExpiredToken, got: %v", err)
	}

	tamperedToken := validToken + "tamper"
	_, err = validator.Validate(tamperedToken)
	if err != ErrInvalidSig {
		t.Fatalf("expected ErrInvalidSig, got: %v", err)
	}
}
