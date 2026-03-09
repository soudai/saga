package daemon

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/soudai/saga/internal/config"
	"github.com/soudai/saga/internal/control"
	"github.com/soudai/saga/internal/runtime"
	"github.com/soudai/saga/internal/store/sqlite"
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

	sqliteStore, err := sqlite.Open(s.paths.DatabasePath())
	if err != nil {
		return err
	}
	defer func() {
		_ = sqliteStore.Close()
	}()

	if err := removeExistingSocket(s.paths.SocketPath); err != nil {
		return err
	}

	listener, err := net.Listen("unix", s.paths.SocketPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = listener.Close()
		_ = os.Remove(s.paths.SocketPath)
	}()
	if err := os.Chmod(s.paths.SocketPath, 0o600); err != nil {
		return err
	}

	httpServer := &http.Server{
		Handler: control.NewServer(sqliteStore).Handler(),
	}

	serveErrCh := make(chan error, 1)
	go func() {
		err := httpServer.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		serveErrCh <- err
	}()

	s.logger.Info("starting saga daemon", "socket_path", s.paths.SocketPath)
	if err := s.notifier.Ready(); err != nil {
		s.logger.Warn("failed to notify readiness; continuing without systemd notification", "error", err)
	}

	select {
	case err := <-serveErrCh:
		if err != nil {
			return err
		}
		return nil
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		_ = httpServer.Close()
		return err
	}
	if err := <-serveErrCh; err != nil {
		return err
	}
	s.logger.Info("stopping saga daemon")
	return nil
}

func removeExistingSocket(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.Mode()&os.ModeSocket == 0 {
		return fmt.Errorf("refusing to remove non-socket path: %s", path)
	}
	return os.Remove(path)
}
