package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/soudai/saga/internal/config"
)

type initProfile string

const (
	initProfileProject initProfile = "project-local"
	initProfileSystem  initProfile = "system-wide"
)

type initAnswers struct {
	ConfigPath string
	StateDir   string
	RunDir     string
	LogDir     string
	SocketPath string
	LogLevel   string
}

type initFileOps struct {
	getwd     func() (string, error)
	stat      func(string) (os.FileInfo, error)
	mkdirAll  func(string, os.FileMode) error
	writeFile func(string, []byte, os.FileMode) error
}

func defaultInitFileOps() initFileOps {
	return initFileOps{
		getwd:     os.Getwd,
		stat:      os.Stat,
		mkdirAll:  os.MkdirAll,
		writeFile: os.WriteFile,
	}
}

func newInitCommand(stdin io.Reader, stdout io.Writer) *cobra.Command {
	return newInitCommandWithOps(stdin, stdout, defaultInitFileOps())
}

func newInitCommandWithOps(stdin io.Reader, stdout io.Writer, ops initFileOps) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init [config-path]",
		Short: "Interactively create an sg config file",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(stdin, stdout, args, force, ops)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing config file without prompting")
	return cmd
}

func runInit(stdin io.Reader, stdout io.Writer, args []string, force bool, ops initFileOps) error {
	wd, err := ops.getwd()
	if err != nil {
		return err
	}

	reader := bufio.NewReader(stdin)
	profile, err := promptProfile(reader, stdout, defaultProfile(args))
	if err != nil {
		return err
	}

	answers := defaultInitAnswers(wd, profile)
	if len(args) > 0 {
		answers.ConfigPath = resolvePath(wd, args[0])
	} else {
		answers.ConfigPath, err = promptValue(reader, stdout, "Config path", answers.ConfigPath)
		if err != nil {
			return err
		}
		answers.ConfigPath = resolvePath(wd, answers.ConfigPath)
	}

	answers.StateDir, err = promptValue(reader, stdout, "State dir", answers.StateDir)
	if err != nil {
		return err
	}
	answers.StateDir = resolvePath(wd, answers.StateDir)
	answers.RunDir, err = promptValue(reader, stdout, "Run dir", answers.RunDir)
	if err != nil {
		return err
	}
	answers.RunDir = resolvePath(wd, answers.RunDir)
	answers.LogDir, err = promptValue(reader, stdout, "Log dir", answers.LogDir)
	if err != nil {
		return err
	}
	answers.LogDir = resolvePath(wd, answers.LogDir)
	answers.SocketPath, err = promptValue(reader, stdout, "Socket path", answers.SocketPath)
	if err != nil {
		return err
	}
	answers.SocketPath = resolvePath(wd, answers.SocketPath)
	answers.LogLevel, err = promptValue(reader, stdout, "Log level", answers.LogLevel)
	if err != nil {
		return err
	}

	cfg := config.Config{
		Runtime: config.RuntimeConfig{
			StateDir: answers.StateDir,
			RunDir:   answers.RunDir,
			LogDir:   answers.LogDir,
		},
		Server: config.ServerConfig{
			SocketPath: answers.SocketPath,
		},
		Log: config.LogConfig{
			Level: answers.LogLevel,
		},
	}

	data, err := config.Marshal(cfg)
	if err != nil {
		return err
	}

	exists, err := fileExists(ops, answers.ConfigPath)
	if err != nil {
		return err
	}
	if exists && !force {
		overwrite, err := promptYesNo(reader, stdout, "Overwrite existing file?", false)
		if err != nil {
			return err
		}
		if !overwrite {
			return fmt.Errorf("config file already exists: %s", answers.ConfigPath)
		}
	}

	if err := ops.mkdirAll(filepath.Dir(answers.ConfigPath), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := ops.writeFile(answers.ConfigPath, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	if err := writef(stdout, "\nWrote %s\n\n", answers.ConfigPath); err != nil {
		return err
	}
	if err := writef(stdout, "Next steps:\n"); err != nil {
		return err
	}
	if err := writef(stdout, "  sg doctor --config %s\n", answers.ConfigPath); err != nil {
		return err
	}
	if err := writef(stdout, "  sg serve --config %s\n", answers.ConfigPath); err != nil {
		return err
	}
	return nil
}

func defaultProfile(args []string) initProfile {
	if len(args) > 0 && strings.HasPrefix(resolvePath("/", args[0]), "/etc/") {
		return initProfileSystem
	}
	return initProfileProject
}

func defaultInitAnswers(wd string, profile initProfile) initAnswers {
	switch profile {
	case initProfileSystem:
		return initAnswers{
			ConfigPath: "/etc/sg/config.yaml",
			StateDir:   "/var/lib/sg",
			RunDir:     "/run/sg",
			LogDir:     "/var/log/sg",
			SocketPath: "/run/sg/sg.sock",
			LogLevel:   "warn",
		}
	default:
		root := filepath.Join(wd, ".sg")
		return initAnswers{
			ConfigPath: filepath.Join(root, "config.yaml"),
			StateDir:   filepath.Join(root, "state"),
			RunDir:     filepath.Join(root, "run"),
			LogDir:     filepath.Join(root, "log"),
			SocketPath: filepath.Join(root, "run", "sg.sock"),
			LogLevel:   "info",
		}
	}
}

func promptProfile(reader *bufio.Reader, stdout io.Writer, profile initProfile) (initProfile, error) {
	if err := writef(stdout, "Select config profile:\n"); err != nil {
		return "", err
	}
	if err := writef(stdout, "  1) project-local\n"); err != nil {
		return "", err
	}
	if err := writef(stdout, "  2) system-wide\n"); err != nil {
		return "", err
	}

	defaultChoice := "1"
	if profile == initProfileSystem {
		defaultChoice = "2"
	}

	for {
		value, err := promptValue(reader, stdout, "Profile", defaultChoice)
		if err != nil {
			return "", err
		}
		switch strings.ToLower(value) {
		case "1", "project", "project-local":
			return initProfileProject, nil
		case "2", "system", "system-wide":
			return initProfileSystem, nil
		default:
			if err := writef(stdout, "Unsupported profile %q. Use 1 or 2.\n", value); err != nil {
				return "", err
			}
		}
	}
}

func promptValue(reader *bufio.Reader, stdout io.Writer, label, defaultValue string) (string, error) {
	if defaultValue == "" {
		if err := writef(stdout, "%s: ", label); err != nil {
			return "", err
		}
	} else {
		if err := writef(stdout, "%s [%s]: ", label, defaultValue); err != nil {
			return "", err
		}
	}

	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}

	value := strings.TrimSpace(line)
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func promptYesNo(reader *bufio.Reader, stdout io.Writer, label string, defaultYes bool) (bool, error) {
	defaultValue := "y/N"
	if defaultYes {
		defaultValue = "Y/n"
	}

	for {
		value, err := promptValue(reader, stdout, label, defaultValue)
		if err != nil {
			return false, err
		}

		switch strings.ToLower(strings.TrimSpace(value)) {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		case "y/n":
			if defaultYes {
				return true, nil
			}
			return false, nil
		default:
			if err := writef(stdout, "Please answer yes or no.\n"); err != nil {
				return false, err
			}
		}
	}
}

func fileExists(ops initFileOps, path string) (bool, error) {
	_, err := ops.stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func resolvePath(wd, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(wd, path)
}

func writef(w io.Writer, format string, args ...any) error {
	_, err := fmt.Fprintf(w, format, args...)
	return err
}
