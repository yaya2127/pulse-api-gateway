package commands

import (
	"fmt"
)

type RouteInfo struct {
	PathPrefix string
	TargetPool string
	CanaryPct  int
}

func ListRoutes() []RouteInfo {
	return []RouteInfo{
		{PathPrefix: "/api/v1/auth", TargetPool: "auth-cluster", CanaryPct: 0},
		{PathPrefix: "/api/v1/payments", TargetPool: "payments-pool", CanaryPct: 10},
		{PathPrefix: "/api/v1/stream", TargetPool: "chronos-nodes", CanaryPct: 5},
	}
}

func PrintRoutes() {
	for _, r := range ListRoutes() {
		fmt.Printf("PREFIX: %-18s -> POOL: %-15s [CANARY: %2d%%]\n", r.PathPrefix, r.TargetPool, r.CanaryPct)
	}
}
