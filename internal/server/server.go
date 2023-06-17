package server

import (
	"context"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"

	"github.com/ivanvc/boombox/internal/config"
	k8s "github.com/ivanvc/boombox/internal/services/kubernetes"
)

// Server holds the boombox server.
type Server struct {
	config *config.Config
	*ssh.Server
	activeSessions sync.WaitGroup

	shuttingDown bool
}

// New returns a new *Server, configured to run boombox.
func New(cfg *config.Config, client *k8s.Client) *Server {
	s := &Server{config: cfg}
	var err error
	s.Server, err = wish.NewServer(
		wish.WithAddress(cfg.Listen),
		wish.WithHostKeyPath(cfg.HostKeyPath),
		wish.WithMiddleware(
			bm.MiddlewareWithProgramHandler(sessionHandler(s, cfg, client), termenv.ANSI256),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("could not start server", "error", err)
		return nil
	}

	return s
}

// Shutdowns the server by closing all active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	s.shuttingDown = true
	if err := s.Close(); err != nil {
		return err
	}
	s.activeSessions.Wait()
	return s.Server.Shutdown(ctx)
}

// Returns true if the server is shutting down.
func (s *Server) IsShuttingDown() bool {
	return s.shuttingDown
}

// Registers a new session.
func (s *Server) RegisterSession() {
	s.activeSessions.Add(1)
}

// Deregisters a session.
func (s *Server) DeregisterSession() {
	s.activeSessions.Done()
}
