package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/C4erries/school/backend/internal/infrastructure/api/config"
)

// Server инкапсулирует http.Server с graceful shutdown.
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
	cfg        *config.Config
}

// NewServer создает новый экземпляр HTTP сервера.
func NewServer(cfg *config.Config, logger *slog.Logger, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.App.Port),
			Handler:      handler,
			ReadTimeout:  cfg.App.Timeout,
			WriteTimeout: cfg.App.Timeout,
			IdleTimeout:  cfg.App.Timeout * 2,
		},
		logger: logger,
		cfg:    cfg,
	}
}

// Start запускает прослушивание входящих HTTP запросов.
func (s *Server) Start() error {
	s.logger.Info("starting HTTP server", slog.String("addr", s.httpServer.Addr))
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server listen: %w", err)
	}
	return nil
}

// Stop плавно останавливает HTTP сервер с учетом таймаута контекста.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server")
	return s.httpServer.Shutdown(ctx)
}

// BuildMux собирает стандартный router на net/http (Go 1.22+ routing patterns).
func BuildMux(logger *slog.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	// Healthcheck endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"service":   "school-api",
		})
	})

	// API v1 prefix healthcheck (для удобства проксирования через nginx)
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"service":   "school-api",
			"version":   "v1",
		})
	})

	return mux
}

// LoggingMiddleware логирует входящие запросы.
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Debug("http request handled",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}
