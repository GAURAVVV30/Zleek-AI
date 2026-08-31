package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hcl-backend/services/api-go/internal/platform/logger"
)

type Server struct {
	httpServer *http.Server
	logger     *logger.Logger
}

func New(port int, handler http.Handler, log *logger.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      handler,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 60 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		logger: log,
	}
}

func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server gracefully...")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to stop server gracefully: %w", err)
	}
	return nil
}
