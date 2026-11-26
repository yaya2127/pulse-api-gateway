package proxy

import (
	"sync"
)

type WeightedNode struct {
	Upstream      *Upstream
	Weight        int
	CurrentWeight int
}

type SmoothWeightedRR struct {
	mu    sync.Mutex
	nodes []*WeightedNode
}

func NewSmoothWeightedRR(nodes []*WeightedNode) *SmoothWeightedRR {
	return &SmoothWeightedRR{nodes: nodes}
}
