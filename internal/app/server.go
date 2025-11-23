package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/111zxc/pr-review-service/internal/config"
	"github.com/111zxc/pr-review-service/internal/logger"
)

type Server struct {
	cfg  *config.Config
	http *http.Server
}

func NewServer(cfg *config.Config, handler http.Handler) *Server {
	srv := &http.Server{
		Addr:         ":" + fmt.Sprint(cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{cfg: cfg, http: srv}
}

func (s *Server) Start() {
	go func() {
		logger.Info("Server is starting", "addr", s.http.Addr)
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", logger.WithError(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.Shutdown()
}

func (s *Server) Shutdown() {
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.http.Shutdown(ctx); err != nil {
		logger.Error("forced shutdown", logger.WithError(err))
	}

	logger.Info("Server stopped")
}
