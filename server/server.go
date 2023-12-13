package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"log/slog"
)

const (
	DefaultMaxHeaderBytes    = 1024
	DefaultReadTimeout       = 1 * time.Second
	DefaultReadHeaderTimeout = 1 * time.Second
	DefaultWriteTimeout      = 2 * time.Second
	DefaultIdleTimeout       = 30 * time.Second
	DefaultAddress           = ""
	DefaultPort              = 3000
	DefaultTelemetryPort     = 9090
)

type IServer interface {
	Start(handler http.Handler) error
	Stop(timeout time.Duration) error
}

// Server is the server part of the application.
type Server struct {
	httpServer        *http.Server
	readTimeout       time.Duration
	readHeaderTimeout time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
	maxHeaderBytes    int
	addr              string
}

// Start starts the server on the provided listener.
func (s *Server) Start(handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:              s.addr,
		Handler:           handler,
		ReadTimeout:       s.readTimeout,
		ReadHeaderTimeout: s.readHeaderTimeout,
		WriteTimeout:      s.writeTimeout,
		IdleTimeout:       s.idleTimeout,
		MaxHeaderBytes:    s.maxHeaderBytes,
	}

	slog.Info("Starting http server", slog.String("address", s.addr))

	if err := s.httpServer.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			return err
		}

		return fmt.Errorf("Server failed to start: %w", err)
	}

	return nil
}

// Stop gracefully stops the server.
func (s *Server) Stop(ctx context.Context) error {
	err := s.httpServer.Shutdown(ctx)
	return err
}

// New create a new server.
func New(rt, rht, wt, it time.Duration, p, mhb int, loc string) *Server {
	return &Server{
		readTimeout:       rt,
		readHeaderTimeout: rht,
		writeTimeout:      wt,
		idleTimeout:       it,
		maxHeaderBytes:    mhb,
		addr:              fmt.Sprintf("%s:%d", loc, p),
	}
}
