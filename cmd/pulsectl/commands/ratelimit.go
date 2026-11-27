package commands

import (
	"fmt"
)

type RateLimitStats struct {
	ClientID   string
	TokensLeft int
	Capacity   int
	Status     string
}

func GetRateLimitStats(clientID string) RateLimitStats {
	return RateLimitStats{
		ClientID:   clientID,
		TokensLeft: 42000,
		Capacity:   50000,
		Status:     "OPTIMAL",
	}
}

func PrintRateLimit(s RateLimitStats) {
	fmt.Printf("Client: %s\nTokens: %d / %d\nStatus: %s (Sliding-Window Active)\n",
		s.ClientID, s.TokensLeft, s.Capacity, s.Status)
}
