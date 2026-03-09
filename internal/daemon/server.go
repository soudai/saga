package daemon

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"

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

	if err := os.RemoveAll(s.paths.SocketPath); err != nil {
		return err
	}

	listener, err := net.Listen("unix", s.paths.SocketPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = listener.Close()
		_ = os.RemoveAll(s.paths.SocketPath)
	}()

	httpServer := &http.Server{
		Handler: control.NewServer(sqliteStore).Handler(),
	}

	var (
		serveErr error
		wg       sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr = err
		}
	}()

	s.logger.Info("starting saga daemon", "socket_path", s.paths.SocketPath)
	if err := s.notifier.Ready(); err != nil {
		return err
	}

	<-ctx.Done()
	if err := httpServer.Shutdown(context.Background()); err != nil {
		return err
	}
	wg.Wait()
	if serveErr != nil {
		return serveErr
	}
	s.logger.Info("stopping saga daemon")
	return nil
}
