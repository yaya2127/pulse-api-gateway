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
