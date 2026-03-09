package runtime

import (
	"os"
	"path/filepath"

	"github.com/soudai/saga/internal/config"
)

type Paths struct {
	StateDir   string
	RunDir     string
	LogDir     string
	SocketPath string
}

func Resolve(cfg config.Config) Paths {
	return Paths{
		StateDir:   cfg.Runtime.StateDir,
		RunDir:     cfg.Runtime.RunDir,
		LogDir:     cfg.Runtime.LogDir,
		SocketPath: cfg.Server.SocketPath,
	}
}

func (p Paths) EnsureDirs() error {
	for _, dir := range []string{
		p.StateDir,
		p.RunDir,
		p.LogDir,
		filepath.Dir(p.SocketPath),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (p Paths) DatabasePath() string {
	return filepath.Join(p.StateDir, "saga.db")
}
