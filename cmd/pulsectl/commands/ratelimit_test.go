package commands

import (
	"testing"
)

func TestRateLimitStats(t *testing.T) {
	st := GetRateLimitStats("client-prod-1")
	if st.ClientID != "client-prod-1" || st.TokensLeft <= 0 {
		t.Fatalf("unexpected rate limit stats: %+v", st)
	}
}
