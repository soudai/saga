package config

import (
	"os"
	"path/filepath"
	"strings"
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

func TestLoadRejectsUnknownField(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("runtime:\n  state_dir: /tmp/saga-state\n  unknown: true\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want unknown field error")
	}
	if !strings.Contains(err.Error(), "field unknown") {
		t.Fatalf("Load() error = %v, want unknown field", err)
	}
}

func TestValidateRejectsRelativeRuntimeDirs(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.Runtime.StateDir = "relative-state"

	if err := Validate(cfg); err == nil || !strings.Contains(err.Error(), "runtime.state_dir must be absolute") {
		t.Fatalf("Validate() error = %v, want absolute path error", err)
	}
}

func TestMarshalRoundTrip(t *testing.T) {
	t.Parallel()

	cfg := Default()
	data, err := Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Server.SocketPath != cfg.Server.SocketPath {
		t.Fatalf("loaded socket path = %q, want %q", loaded.Server.SocketPath, cfg.Server.SocketPath)
	}
	if loaded.Log.Level != cfg.Log.Level {
		t.Fatalf("loaded log level = %q, want %q", loaded.Log.Level, cfg.Log.Level)
	}
}
