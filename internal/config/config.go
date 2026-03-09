package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Runtime RuntimeConfig `yaml:"runtime"`
	Server  ServerConfig  `yaml:"server"`
	Log     LogConfig     `yaml:"log"`
}

type RuntimeConfig struct {
	StateDir string `yaml:"state_dir"`
	RunDir   string `yaml:"run_dir"`
	LogDir   string `yaml:"log_dir"`
}

type ServerConfig struct {
	SocketPath string `yaml:"socket_path"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

func Default() Config {
	return Config{
		Runtime: RuntimeConfig{
			StateDir: "/var/lib/saga",
			RunDir:   "/run/saga",
			LogDir:   "/var/log/saga",
		},
		Server: ServerConfig{
			SocketPath: "/run/saga/saga.sock",
		},
		Log: LogConfig{
			Level: "info",
		},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()

	if path != "" {
		raw, err := os.ReadFile(path)
		if err != nil {
			return Config{}, fmt.Errorf("read config: %w", err)
		}

		var fileCfg Config
		if err := yaml.Unmarshal(raw, &fileCfg); err != nil {
			return Config{}, fmt.Errorf("parse config: %w", err)
		}

		merge(&cfg, fileCfg)
	}

	applyEnv(&cfg)
	if err := Validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func Validate(cfg Config) error {
	if cfg.Runtime.StateDir == "" {
		return fmt.Errorf("runtime.state_dir is required")
	}
	if cfg.Runtime.RunDir == "" {
		return fmt.Errorf("runtime.run_dir is required")
	}
	if cfg.Runtime.LogDir == "" {
		return fmt.Errorf("runtime.log_dir is required")
	}
	if cfg.Server.SocketPath == "" {
		return fmt.Errorf("server.socket_path is required")
	}
	if !filepath.IsAbs(cfg.Server.SocketPath) {
		return fmt.Errorf("server.socket_path must be absolute")
	}
	return nil
}

func merge(dst *Config, src Config) {
	if src.Runtime.StateDir != "" {
		dst.Runtime.StateDir = src.Runtime.StateDir
	}
	if src.Runtime.RunDir != "" {
		dst.Runtime.RunDir = src.Runtime.RunDir
	}
	if src.Runtime.LogDir != "" {
		dst.Runtime.LogDir = src.Runtime.LogDir
	}
	if src.Server.SocketPath != "" {
		dst.Server.SocketPath = src.Server.SocketPath
	}
	if src.Log.Level != "" {
		dst.Log.Level = src.Log.Level
	}
}

func applyEnv(cfg *Config) {
	override(&cfg.Runtime.StateDir, "SAGA_STATE_DIR")
	override(&cfg.Runtime.RunDir, "SAGA_RUN_DIR")
	override(&cfg.Runtime.LogDir, "SAGA_LOG_DIR")
	override(&cfg.Server.SocketPath, "SAGA_SOCKET_PATH")
	override(&cfg.Log.Level, "SAGA_LOG_LEVEL")
}

func override(dst *string, key string) {
	if value := os.Getenv(key); value != "" {
		*dst = value
	}
}
