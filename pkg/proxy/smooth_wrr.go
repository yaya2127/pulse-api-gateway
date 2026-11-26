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

func (s *SmoothWeightedRR) Next() *Upstream {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.nodes) == 0 {
		return nil
	}

	totalWeight := 0
	var best *WeightedNode

	for _, node := range s.nodes {
		if !node.Upstream.Healthy {
			continue
		}
		node.CurrentWeight += node.Weight
		totalWeight += node.Weight

		if best == nil || node.CurrentWeight > best.CurrentWeight {
			best = node
		}
	}

	if best == nil {
		return nil
	}

	best.CurrentWeight -= totalWeight
	return best.Upstream
}
