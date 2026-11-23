package middleware

import (
	"testing"
	"time"
)

func BenchmarkJWTValidation(b *testing.B) {
	secret := "benchmark-super-secret-key"
	validator := NewJWTValidator(secret)
	token := makeTestToken(secret, time.Hour)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = validator.Validate(token)
		}
	})
}
