package doctor

import (
	"path/filepath"

	"github.com/soudai/saga/internal/config"
)

type Check struct {
	Name   string
	OK     bool
	Detail string
}

func Run(cfg config.Config) []Check {
	return []Check{
		{
			Name:   "config validation",
			OK:     true,
			Detail: "configuration loaded successfully",
		},
		{
			Name:   "socket path",
			OK:     filepath.IsAbs(cfg.Server.SocketPath),
			Detail: cfg.Server.SocketPath,
		},
		{
			Name:   "runtime state dir",
			OK:     filepath.IsAbs(cfg.Runtime.StateDir),
			Detail: cfg.Runtime.StateDir,
		},
		{
			Name:   "runtime run dir",
			OK:     filepath.IsAbs(cfg.Runtime.RunDir),
			Detail: cfg.Runtime.RunDir,
		},
		{
			Name:   "runtime log dir",
			OK:     filepath.IsAbs(cfg.Runtime.LogDir),
			Detail: cfg.Runtime.LogDir,
		},
	}
}
