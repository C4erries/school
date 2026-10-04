package di

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver"
	"github.com/C4erries/school/backend/internal/infrastructure/api/config"
)

// Container объединяет все зависимости API сервиса (DI сборка).
type Container struct {
	Config     *config.Config
	Logger     *slog.Logger
	HTTPServer *httpserver.Server
}

// NewContainer инициализирует все адаптеры и зависимости согласно конфигурации.
func NewContainer(cfg *config.Config) (*Container, error) {
	// Инициализация логгера slog
	var level slog.Level
	switch cfg.App.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	// Сборка HTTP роутера и middleware
	mux := httpserver.BuildMux(logger)
	handlerWithLogging := httpserver.LoggingMiddleware(logger)(mux)

	// Создание HTTP сервера
	server := httpserver.NewServer(cfg, logger, handlerWithLogging)

	return &Container{
		Config:     cfg,
		Logger:     logger,
		HTTPServer: server,
	}, nil
}

// Close освобождает открытые ресурсы контейнера (БД, кэш и т.д.).
func (c *Container) Close() error {
	c.Logger.Info("closing container resources")
	var errs []error

	// В будущем здесь закрываем connection pool Postgres, Valkey клиент и т.д.
	if len(errs) > 0 {
		return fmt.Errorf("errors closing container: %v", errs)
	}
	return nil
}
