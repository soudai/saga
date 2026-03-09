package doctor

import (
	"os"
	"path/filepath"
	"strings"

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
			Name:   "systemd available",
			OK:     hasSystemd(),
			Detail: "/run/systemd/system",
		},
		{
			Name:   "wsl2 environment",
			OK:     isWSL(),
			Detail: "detected from /proc/version or WSL_INTEROP",
		},
		{
			Name:   "linux filesystem recommendation",
			OK:     !strings.HasPrefix(cfg.Runtime.StateDir, "/mnt/"),
			Detail: cfg.Runtime.StateDir,
		},
	}
}

func hasSystemd() bool {
	_, err := os.Stat("/run/systemd/system")
	return err == nil
}

func isWSL() bool {
	if os.Getenv("WSL_INTEROP") != "" {
		return true
	}

	raw, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(raw)), "microsoft")
}
