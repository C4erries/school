package di

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	valkeylib "github.com/valkey-io/valkey-go"

	"github.com/C4erries/school/backend/internal/application/auth"
	"github.com/C4erries/school/backend/internal/application/crm"
	"github.com/C4erries/school/backend/internal/application/dashboard"
	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/postgres"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/security"
	valkeyadapter "github.com/C4erries/school/backend/internal/infrastructure/api/adapters/valkey"
	"github.com/C4erries/school/backend/internal/infrastructure/api/config"
)

// Container объединяет все зависимости API сервиса (DI сборка).
type Container struct {
	Config           *config.Config
	Logger           *slog.Logger
	DB               *sql.DB
	ValkeyClient     valkeylib.Client
	HTTPServer       *httpserver.Server
	AuthService      *auth.Service
	ScheduleService  *schedule.Service
	CRMService       *crm.Service
	DashboardService *dashboard.Service
}

// NewContainer инициализирует все адаптеры и зависимости согласно конфигурации.
func NewContainer(cfg *config.Config) (*Container, error) {
	// 1. Инициализация логгера slog
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

	// 2. Подключение к PostgreSQL
	db, err := postgres.NewConnection(cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("init postgres connection: %w", err)
	}

	// 3. Подключение к Valkey
	valkeyClient, err := valkeylib.NewClient(valkeylib.ClientOption{
		InitAddress: []string{cfg.Valkey.Addr()},
		Password:    cfg.Valkey.Password,
	})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("init valkey client: %w", err)
	}

	// 4. Репозитории и адаптеры инфраструктуры
	_ = postgres.NewTransactor(db)
	userRepo := postgres.NewUserRepository(db)
	classroomRepo := postgres.NewClassroomRepository(db)
	clientRepo := postgres.NewClientRepository(db)
	subRepo := postgres.NewSubscriptionRepository(db)
	tagRepo := postgres.NewTagRepository(db)
	lessonRepo := postgres.NewLessonRepository(db)
	passwordHasher := security.NewPasswordHasher(12)
	tokenManager := security.NewTokenManager(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	sessionStore := valkeyadapter.NewSessionStore(valkeyClient)

	// 5. Сервисы уровня Application
	authService := auth.NewService(userRepo, passwordHasher, tokenManager, sessionStore)
	scheduleService := schedule.NewService(classroomRepo, lessonRepo, clientRepo, subRepo)
	crmService := crm.NewService(clientRepo, subRepo, tagRepo)
	dashboardService := dashboard.NewService(lessonRepo, clientRepo)

	// 6. HTTP API Handler и роутер
	apiHandler := httpserver.NewAPIHandler(
		authService,
		scheduleService,
		crmService,
		dashboardService,
		tokenManager,
		cfg.App.Version,
	)
	mux := httpserver.BuildMux(apiHandler, logger)
	handlerWithLogging := httpserver.LoggingMiddleware(logger)(mux)

	// 7. Создание HTTP сервера
	server := httpserver.NewServer(cfg, logger, handlerWithLogging)

	return &Container{
		Config:           cfg,
		Logger:           logger,
		DB:               db,
		ValkeyClient:     valkeyClient,
		HTTPServer:       server,
		AuthService:      authService,
		ScheduleService:  scheduleService,
		CRMService:       crmService,
		DashboardService: dashboardService,
	}, nil
}

// Close освобождает открытые ресурсы контейнера (БД, кэш и т.д.).
func (c *Container) Close() error {
	c.Logger.Info("closing container resources")
	var errs []error

	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close postgres db: %w", err))
		}
	}

	if c.ValkeyClient != nil {
		c.ValkeyClient.Close()
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing container: %v", errs)
	}
	return nil
}
