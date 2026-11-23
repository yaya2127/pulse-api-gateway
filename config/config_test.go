package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.json")
	content := []byte(`{
		"server": { "port": 8888, "metrics_port": 9999 },
		"routes": [
			{ "path_prefix": "/test", "upstreams": [{"url": "http://localhost:3000", "weight": 1}] }
		]
	}`)

	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Server.Port != 8888 {
		t.Errorf("expected port 8888, got %d", cfg.Server.Port)
	}
	if len(cfg.Routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(cfg.Routes))
	}
}
