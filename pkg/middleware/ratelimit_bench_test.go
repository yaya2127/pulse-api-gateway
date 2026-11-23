package middleware

import (
	"testing"
)

func BenchmarkTokenBucketAllow(b *testing.B) {
	tb := NewTokenBucket(1000000, 100000)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tb.Allow()
		}
	})
}
