package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefault(t *testing.T) {
	t.Parallel()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.SocketPath != "/run/saga/saga.sock" {
		t.Fatalf("socket path = %q, want default", cfg.Server.SocketPath)
	}
}

func TestLoadFileOverride(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("runtime:\n  state_dir: /tmp/saga-state\nserver:\n  socket_path: /tmp/saga.sock\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Runtime.StateDir != "/tmp/saga-state" {
		t.Fatalf("state dir = %q, want override", cfg.Runtime.StateDir)
	}
	if cfg.Server.SocketPath != "/tmp/saga.sock" {
		t.Fatalf("socket path = %q, want override", cfg.Server.SocketPath)
	}
}
