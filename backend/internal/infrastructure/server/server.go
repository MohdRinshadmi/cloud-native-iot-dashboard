// Package server wraps net/http with production-grade timeouts and graceful
// shutdown. Separating this from routing means the transport can be tested and
// reasoned about independently of the routes it serves.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/ioss/iot-dashboard/backend/internal/infrastructure/config"
)

// Server is a thin lifecycle wrapper around *http.Server.
type Server struct {
	httpServer *http.Server
	log        *slog.Logger
}

// New constructs the HTTP server with timeouts sourced from config. Timeouts
// are mandatory in production: without them a slow client can exhaust the
// connection pool (Slowloris).
func New(cfg config.HTTPConfig, handler http.Handler, log *slog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.Addr(),
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.ReadTimeout * 4,
		},
		log: log,
	}
}

// Start begins serving and blocks until the server stops. A clean shutdown
// returns nil (http.ErrServerClosed is expected, not an error).
func (s *Server) Start() error {
	s.log.Info("http server listening", slog.String("addr", s.httpServer.Addr))
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown drains in-flight requests, bounded by ctx. Call this when a SIGINT/
// SIGTERM is received so Kubernetes rolling deploys don't drop connections.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("http server shutting down")
	return s.httpServer.Shutdown(ctx)
}
