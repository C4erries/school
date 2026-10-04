package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/C4erries/school/backend/internal/infrastructure/api/config"
	"github.com/C4erries/school/backend/internal/infrastructure/api/di"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application exited with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	// 1. Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// 2. Инициализация DI контейнера
	container, err := di.NewContainer(cfg)
	if err != nil {
		return fmt.Errorf("init container: %w", err)
	}
	defer func() {
		if closeErr := container.Close(); closeErr != nil {
			container.Logger.Error("failed to close container", slog.String("error", closeErr.Error()))
		}
	}()

	container.Logger.Info("initialized application",
		slog.String("env", cfg.App.Env),
		slog.Int("port", cfg.App.Port),
	)

	// 3. Канал для перехвата сигналов ОС (graceful shutdown)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 4. Запуск HTTP сервера в отдельной горутине
	serverErrChan := make(chan error, 1)
	go func() {
		if err := container.HTTPServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
	}()

	// 5. Ожидание сигнала остановки или фатальной ошибки сервера
	select {
	case sig := <-sigChan:
		container.Logger.Info("received termination signal", slog.String("signal", sig.String()))
	case err := <-serverErrChan:
		return fmt.Errorf("http server failure: %w", err)
	}

	// 6. Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := container.HTTPServer.Stop(shutdownCtx); err != nil {
		return fmt.Errorf("stop http server: %w", err)
	}

	container.Logger.Info("application stopped gracefully")
	return nil
}
