package cli

import (
	"context"
	"fmt"
	"io"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/soudai/saga/internal/config"
	"github.com/soudai/saga/internal/control"
	"github.com/soudai/saga/internal/daemon"
	"github.com/soudai/saga/internal/doctor"
	"github.com/soudai/saga/internal/logging"
	"github.com/soudai/saga/internal/runtime"
	sagasystemd "github.com/soudai/saga/internal/systemd"
	"github.com/soudai/saga/internal/version"
)

func NewRootCommand(stdin io.Reader, stdout, stderr io.Writer) *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:           "saga",
		Short:         "Saga AI agent framework",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.SetIn(stdin)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	cmd.PersistentFlags().StringVar(&configPath, "config", "", "path to config file")

	cmd.AddCommand(newInitCommand(stdin, stdout))
	cmd.AddCommand(newVersionCommand(stdout))
	cmd.AddCommand(newDoctorCommand(stdout, stderr, &configPath))
	cmd.AddCommand(newServeCommand(stdout, stderr, &configPath))
	cmd.AddCommand(newStatusCommand(stdout, &configPath))
	cmd.AddCommand(newTaskActionCommand("cancel", "Cancel a task", &configPath))
	cmd.AddCommand(newTaskActionCommand("retry", "Retry a task", &configPath))
	cmd.AddCommand(newTaskActionCommand("resume", "Resume a task", &configPath))
	return cmd
}

func newVersionCommand(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print build information",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(stdout, version.String())
			return err
		},
	}
}

func newDoctorCommand(stdout, stderr io.Writer, configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run local environment checks",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			checks := doctor.Run(cfg)
			for _, check := range checks {
				status := "OK"
				if !check.OK {
					status = "NG"
				}
				if _, err := fmt.Fprintf(stdout, "%s\t%s\t%s\n", status, check.Name, check.Detail); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func newServeCommand(stdout, stderr io.Writer, configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Run the saga daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			logger, err := logging.New(cfg.Log.Level, stderr)
			if err != nil {
				return err
			}

			paths := runtime.Resolve(cfg)
			server := daemon.New(cfg, logger, paths, sagasystemd.New())

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			return server.Serve(ctx)
		},
	}
}

func newStatusCommand(stdout io.Writer, configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show daemon status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			status, err := control.NewClient(cfg.Server.SocketPath).Status(cmd.Context())
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintf(stdout, "tasks=%d active_runs=%d\n", len(status.Tasks), status.ActiveRuns); err != nil {
				return err
			}
			for _, task := range status.Tasks {
				if _, err := fmt.Fprintf(stdout, "task id=%d repo=%s issue=%d state=%s\n", task.ID, task.Repository, task.IssueNumber, task.State); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func newTaskActionCommand(action, short string, configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   action + " <task-id>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			taskID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task id: %w", err)
			}

			client := control.NewClient(cfg.Server.SocketPath)
			switch action {
			case "cancel":
				return client.Cancel(cmd.Context(), taskID)
			case "retry":
				return client.Retry(cmd.Context(), taskID)
			case "resume":
				return client.Resume(cmd.Context(), taskID)
			default:
				return fmt.Errorf("unsupported action: %s", action)
			}
		},
	}
}
