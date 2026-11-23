package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

func main() {
	secret := []byte("pulse-secure-secret-key-2026")
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	claims := map[string]any{
		"sub":  "usr_981247",
		"iss":  "pulse-gateway",
		"aud":  "cloud-api",
		"role": "engineer",
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	claimsBytes, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(claimsBytes)

	hp := header + "." + payload
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(hp))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	token := hp + "." + sig
	fmt.Printf("Generated Test Token:\n%s\n", token)
}
