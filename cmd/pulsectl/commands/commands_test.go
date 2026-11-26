package commands

import (
	"testing"
)

func TestCommands(t *testing.T) {
	st := GetStatus()
	if st.Version != "v1.1.0" {
		t.Errorf("got version %s, want v1.1.0", st.Version)
	}

	dm := NewDrainManager()
	if err := dm.DrainNode("worker-01"); err != nil {
		t.Errorf("failed to drain node: %v", err)
	}

	routes := ListRoutes()
	if len(routes) == 0 {
		t.Fatal("expected non-empty routes list")
	}
}
