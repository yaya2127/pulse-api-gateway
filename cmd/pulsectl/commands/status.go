package commands

import (
	"fmt"
	"runtime"
)

type GatewayStatus struct {
	Version     string
	ActiveConns int
	Goroutines  int
	HeapAllocMB float64
}

func GetStatus() GatewayStatus {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return GatewayStatus{
		Version:     "v1.1.0",
		ActiveConns: 14250,
		Goroutines:  runtime.NumGoroutine(),
		HeapAllocMB: float64(m.Alloc) / 1024 / 1024,
	}
}

func FormatStatus(s GatewayStatus) string {
	return fmt.Sprintf("Pulse API Gateway %s\nActive Connections: %d\nGoroutines: %d\nHeap: %.2f MB",
		s.Version, s.ActiveConns, s.Goroutines, s.HeapAllocMB)
}
