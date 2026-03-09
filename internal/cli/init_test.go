package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/soudai/saga/internal/config"
)

func TestInitCommandCreatesProjectLocalConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	var stdout bytes.Buffer

	cmd := newInitCommandWithOps(strings.NewReader("\n\n\n\n\n\n"), &stdout, initFileOps{
		getwd:     func() (string, error) { return dir, nil },
		stat:      os.Stat,
		mkdirAll:  os.MkdirAll,
		writeFile: os.WriteFile,
	})
	cmd.SetArgs(nil)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	configPath := filepath.Join(dir, ".saga", "config.yaml")
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.SocketPath != filepath.Join(dir, ".saga", "run", "saga.sock") {
		t.Fatalf("socket path = %q, want project-local socket", cfg.Server.SocketPath)
	}
	if !strings.Contains(stdout.String(), "saga serve --config "+configPath) {
		t.Fatalf("stdout = %q, want next steps output", stdout.String())
	}
}

func TestInitCommandResolvesRelativePaths(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cmd := newInitCommandWithOps(strings.NewReader("\n\nstate\nrun\nlog\nsocket.sock\n\n"), io.Discard, initFileOps{
		getwd:     func() (string, error) { return dir, nil },
		stat:      os.Stat,
		mkdirAll:  os.MkdirAll,
		writeFile: os.WriteFile,
	})
	cmd.SetArgs(nil)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	configPath := filepath.Join(dir, ".saga", "config.yaml")
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Runtime.StateDir != filepath.Join(dir, "state") {
		t.Fatalf("state dir = %q, want %q", cfg.Runtime.StateDir, filepath.Join(dir, "state"))
	}
	if cfg.Runtime.RunDir != filepath.Join(dir, "run") {
		t.Fatalf("run dir = %q, want %q", cfg.Runtime.RunDir, filepath.Join(dir, "run"))
	}
	if cfg.Runtime.LogDir != filepath.Join(dir, "log") {
		t.Fatalf("log dir = %q, want %q", cfg.Runtime.LogDir, filepath.Join(dir, "log"))
	}
	if cfg.Server.SocketPath != filepath.Join(dir, "socket.sock") {
		t.Fatalf("socket path = %q, want %q", cfg.Server.SocketPath, filepath.Join(dir, "socket.sock"))
	}
}

func TestInitCommandUsesSystemProfile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "etc", "saga", "config.yaml")

	cmd := newInitCommandWithOps(strings.NewReader("2\n\n\n\n\n\n"), io.Discard, initFileOps{
		getwd:     func() (string, error) { return dir, nil },
		stat:      os.Stat,
		mkdirAll:  os.MkdirAll,
		writeFile: os.WriteFile,
	})
	cmd.SetArgs([]string{target})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	cfg, err := config.Load(target)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Runtime.StateDir != "/var/lib/saga" {
		t.Fatalf("state dir = %q, want system default", cfg.Runtime.StateDir)
	}
	if cfg.Log.Level != "warn" {
		t.Fatalf("log level = %q, want warn", cfg.Log.Level)
	}
}

func TestInitCommandPromptsBeforeOverwrite(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, ".saga", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(configPath, []byte("existing"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	cmd := newInitCommandWithOps(strings.NewReader("\n\n\n\n\n\nn\n"), &stdout, initFileOps{
		getwd:     func() (string, error) { return dir, nil },
		stat:      os.Stat,
		mkdirAll:  os.MkdirAll,
		writeFile: os.WriteFile,
	})
	cmd.SetArgs(nil)

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("Execute() error = %v, want overwrite refusal", err)
	}
}

func TestInitCommandReturnsWriteError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cmd := newInitCommandWithOps(strings.NewReader(""), errWriter{}, initFileOps{
		getwd:     func() (string, error) { return dir, nil },
		stat:      os.Stat,
		mkdirAll:  os.MkdirAll,
		writeFile: os.WriteFile,
	})
	cmd.SetArgs(nil)

	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want write error")
	}
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) {
	return 0, io.ErrClosedPipe
}
