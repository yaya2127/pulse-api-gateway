package commands

import (
	"errors"
	"fmt"
)

type DrainManager struct {
	drainedNodes map[string]bool
}

func NewDrainManager() *DrainManager {
	return &DrainManager{drainedNodes: make(map[string]bool)}
}

func (d *DrainManager) DrainNode(nodeID string) error {
	if nodeID == "" {
		return errors.New("node ID cannot be empty")
	}
	d.drainedNodes[nodeID] = true
	fmt.Printf("[DRAIN] Node %s marked for zero-downtime drain\n", nodeID)
	return nil
}
