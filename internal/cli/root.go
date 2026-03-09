package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/soudai/saga/internal/config"
	"github.com/soudai/saga/internal/daemon"
	"github.com/soudai/saga/internal/doctor"
	"github.com/soudai/saga/internal/logging"
	"github.com/soudai/saga/internal/runtime"
	sagasystemd "github.com/soudai/saga/internal/systemd"
	"github.com/soudai/saga/internal/version"
)

var ErrPrinted = errors.New("message already written to output")

func NewRootCommand(stdout, stderr io.Writer) *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:           "saga",
		Short:         "Saga AI agent framework",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&configPath, "config", "", "path to config file")

	cmd.AddCommand(newVersionCommand(stdout))
	cmd.AddCommand(newDoctorCommand(stdout, stderr, &configPath))
	cmd.AddCommand(newServeCommand(stdout, stderr, &configPath))
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
