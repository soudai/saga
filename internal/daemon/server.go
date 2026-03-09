package daemon

import (
	"context"
	"log/slog"

	"github.com/soudai/saga/internal/config"
	"github.com/soudai/saga/internal/runtime"
	sagasystemd "github.com/soudai/saga/internal/systemd"
)

type Server struct {
	config   config.Config
	logger   *slog.Logger
	paths    runtime.Paths
	notifier sagasystemd.Notifier
}

func New(cfg config.Config, logger *slog.Logger, paths runtime.Paths, notifier sagasystemd.Notifier) *Server {
	return &Server{
		config:   cfg,
		logger:   logger,
		paths:    paths,
		notifier: notifier,
	}
}

func (s *Server) Serve(ctx context.Context) error {
	if err := s.paths.EnsureDirs(); err != nil {
		return err
	}

	s.logger.Info("starting saga daemon", "socket_path", s.paths.SocketPath)
	if err := s.notifier.Ready(); err != nil {
		return err
	}

	<-ctx.Done()
	s.logger.Info("stopping saga daemon")
	return nil
}
