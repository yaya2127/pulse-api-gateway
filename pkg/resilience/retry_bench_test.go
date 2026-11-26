package resilience

import (
	"context"
	"net/http"
	"testing"
)

func BenchmarkRetryPolicy_ImmediateSuccess(b *testing.B) {
	policy := DefaultRetryPolicy()
	ctx := context.Background()
	okResp := &http.Response{StatusCode: http.StatusOK}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = policy.Execute(ctx, func() (*http.Response, error) {
			return okResp, nil
		})
	}
}
