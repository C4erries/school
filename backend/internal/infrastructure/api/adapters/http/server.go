package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/C4erries/school/backend/internal/application/analytics"
	"github.com/C4erries/school/backend/internal/application/auth"
	"github.com/C4erries/school/backend/internal/application/calendar"
	"github.com/C4erries/school/backend/internal/application/crm"
	"github.com/C4erries/school/backend/internal/application/dashboard"
	"github.com/C4erries/school/backend/internal/application/finance"
	"github.com/C4erries/school/backend/internal/application/schedule"
	httpanalytics "github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/analytics"
	httpauth "github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/auth"
	httpcrm "github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/crm"
	httpdashboard "github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/dashboard"
	httpfinance "github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/finance"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/middleware"
	httpschedule "github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/schedule"
	"github.com/C4erries/school/backend/internal/infrastructure/api/config"
)

// Server реализует сгенерированный generated.ServerInterface и объединяет доменные хэндлеры Echo.
type Server struct {
	echoInstance *echo.Echo
	logger       *slog.Logger
	cfg          *config.Config
	version      string

	authHandler         *httpauth.Handler
	clientHandler       *httpcrm.ClientHandler
	subscriptionHandler *httpcrm.SubscriptionHandler
	tagHandler          *httpcrm.TagHandler
	lessonHandler       *httpschedule.LessonHandler
	classroomHandler    *httpschedule.ClassroomHandler
	calendarHandler     *httpschedule.CalendarHandler
	seriesHandler       *httpschedule.SeriesHandler
	financeHandler      *httpfinance.FinanceHandler
	exportHandler       *httpfinance.ExportHandler
	analyticsHandler    *httpanalytics.AnalyticsHandler
	dashboardHandler    *httpdashboard.DashboardHandler
}

// NewServer создает новый экземпляр HTTP сервера и настраивает маршрутизацию.
func NewServer(
	cfg *config.Config,
	logger *slog.Logger,
	authService *auth.Service,
	scheduleService *schedule.Service,
	crmService *crm.Service,
	dashboardService *dashboard.Service,
	tokenValidator middleware.TokenValidator,
	version string,
	extra ...any,
) *Server {
	var financeService *finance.Service
	var analyticsService *analytics.Service
	var calendarService *calendar.Service

	for _, ext := range extra {
		if f, ok := ext.(*finance.Service); ok {
			financeService = f
		}
		if a, ok := ext.(*analytics.Service); ok {
			analyticsService = a
		}
		if c, ok := ext.(*calendar.Service); ok {
			calendarService = c
		}
	}

	authH := httpauth.NewHandler(authService, tokenValidator)
	clientH := httpcrm.NewClientHandler(crmService, authH)
	subH := httpcrm.NewSubscriptionHandler(crmService, authH)
	tagH := httpcrm.NewTagHandler(crmService, authH)
	lessonH := httpschedule.NewLessonHandler(scheduleService, authH)
	classroomH := httpschedule.NewClassroomHandler(scheduleService, authH)
	calendarH := httpschedule.NewCalendarHandler(calendarService, authH)
	seriesH := httpschedule.NewSeriesHandler(scheduleService, authH)
	financeH := httpfinance.NewFinanceHandler(financeService, authH)
	exportH := httpfinance.NewExportHandler(financeService, authH)
	analyticsH := httpanalytics.NewAnalyticsHandler(analyticsService, authH)
	dashboardH := httpdashboard.NewDashboardHandler(dashboardService, authH)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Standard middleware
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{"*"},
	}))
	e.Use(middleware.RequestLogger(logger))

	s := &Server{
		echoInstance:        e,
		logger:              logger,
		cfg:                 cfg,
		version:             version,
		authHandler:         authH,
		clientHandler:       clientH,
		subscriptionHandler: subH,
		tagHandler:          tagH,
		lessonHandler:       lessonH,
		classroomHandler:    classroomH,
		calendarHandler:     calendarH,
		seriesHandler:       seriesH,
		financeHandler:      financeH,
		exportHandler:       exportH,
		analyticsHandler:    analyticsH,
		dashboardHandler:    dashboardH,
	}

	// Register generated handlers directly and with /api/v1 prefix
	generated.RegisterHandlers(e, s)
	generated.RegisterHandlersWithBaseURL(e, s, "/api/v1")

	return s
}

// Echo возвращает базовый экземпляр echo.Echo.
func (s *Server) Echo() *echo.Echo {
	return s.echoInstance
}

// Start запускает сервер на указанном порту.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.App.Port)
	s.logger.Info("starting Echo HTTP server", slog.String("addr", addr))

	s.echoInstance.Server.ReadTimeout = s.cfg.App.Timeout
	s.echoInstance.Server.WriteTimeout = s.cfg.App.Timeout
	s.echoInstance.Server.IdleTimeout = s.cfg.App.Timeout * 2

	if err := s.echoInstance.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("echo server listen: %w", err)
	}
	return nil
}

// Stop плавно останавливает HTTP сервер.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("shutting down Echo HTTP server")
	return s.echoInstance.Shutdown(ctx)
}

// Health check endpoint
func (s *Server) GetHealth(ctx echo.Context) error {
	resp := generated.HealthResponse{
		Status:    "ok",
		Service:   "school-api",
		Timestamp: time.Now().UTC(),
		Version:   &s.version,
	}
	return ctx.JSON(http.StatusOK, resp)
}

// Auth delegates
func (s *Server) Register(ctx echo.Context) error      { return s.authHandler.Register(ctx) }
func (s *Server) Login(ctx echo.Context) error         { return s.authHandler.Login(ctx) }
func (s *Server) RefreshTokens(ctx echo.Context) error { return s.authHandler.RefreshTokens(ctx) }
func (s *Server) GetCurrentUser(ctx echo.Context) error {
	return s.authHandler.GetCurrentUser(ctx)
}
func (s *Server) GetUserDefaultRates(ctx echo.Context) error {
	return s.authHandler.GetUserDefaultRates(ctx)
}
func (s *Server) UpdateUserDefaultRates(ctx echo.Context) error {
	return s.authHandler.UpdateUserDefaultRates(ctx)
}

// CRM delegates
func (s *Server) ListClients(ctx echo.Context, params generated.ListClientsParams) error {
	return s.clientHandler.ListClients(ctx, params)
}
func (s *Server) CreateClient(ctx echo.Context) error { return s.clientHandler.CreateClient(ctx) }
func (s *Server) GetClient(ctx echo.Context, id openapi_types.UUID) error {
	return s.clientHandler.GetClient(ctx, id)
}
func (s *Server) UpdateClient(ctx echo.Context, id openapi_types.UUID) error {
	return s.clientHandler.UpdateClient(ctx, id)
}
func (s *Server) DeleteClient(ctx echo.Context, id openapi_types.UUID) error {
	return s.clientHandler.DeleteClient(ctx, id)
}
func (s *Server) ArchiveClient(ctx echo.Context, id openapi_types.UUID) error {
	return s.clientHandler.ArchiveClient(ctx, id)
}
func (s *Server) UnarchiveClient(ctx echo.Context, id openapi_types.UUID) error {
	return s.clientHandler.UnarchiveClient(ctx, id)
}
func (s *Server) ListSubscriptions(ctx echo.Context, id openapi_types.UUID) error {
	return s.subscriptionHandler.ListSubscriptions(ctx, id)
}
func (s *Server) CreateSubscription(ctx echo.Context, id openapi_types.UUID) error {
	return s.subscriptionHandler.CreateSubscription(ctx, id)
}
func (s *Server) AdjustClientBalance(ctx echo.Context, id openapi_types.UUID) error {
	return s.subscriptionHandler.AdjustClientBalance(ctx, id)
}
func (s *Server) ListTags(ctx echo.Context) error  { return s.tagHandler.ListTags(ctx) }
func (s *Server) CreateTag(ctx echo.Context) error { return s.tagHandler.CreateTag(ctx) }
func (s *Server) DeleteTag(ctx echo.Context, id openapi_types.UUID) error {
	return s.tagHandler.DeleteTag(ctx, id)
}
func (s *Server) AssignClientTag(ctx echo.Context, id openapi_types.UUID) error {
	return s.tagHandler.AssignClientTag(ctx, id)
}
func (s *Server) RemoveClientTag(ctx echo.Context, id openapi_types.UUID, tagID openapi_types.UUID) error {
	return s.tagHandler.RemoveClientTag(ctx, id, tagID)
}

// Schedule & Calendar delegates
func (s *Server) ListClassrooms(ctx echo.Context) error {
	return s.classroomHandler.ListClassrooms(ctx)
}
func (s *Server) CreateClassroom(ctx echo.Context) error {
	return s.classroomHandler.CreateClassroom(ctx)
}
func (s *Server) ListLessons(ctx echo.Context, params generated.ListLessonsParams) error {
	return s.lessonHandler.ListLessons(ctx, params)
}
func (s *Server) CreateLesson(ctx echo.Context) error { return s.lessonHandler.CreateLesson(ctx) }
func (s *Server) CompleteLesson(ctx echo.Context, id openapi_types.UUID) error {
	return s.lessonHandler.CompleteLesson(ctx, id)
}
func (s *Server) CancelLesson(ctx echo.Context, id openapi_types.UUID) error {
	return s.lessonHandler.CancelLesson(ctx, id)
}
func (s *Server) UpdateLesson(ctx echo.Context, id openapi_types.UUID) error {
	return s.lessonHandler.UpdateLesson(ctx, id)
}
func (s *Server) GetCalendarSettings(ctx echo.Context) error {
	return s.calendarHandler.GetCalendarSettings(ctx)
}
func (s *Server) RotateCalendarToken(ctx echo.Context) error {
	return s.calendarHandler.RotateCalendarToken(ctx)
}
func (s *Server) ExportCalendarFile(ctx echo.Context, params generated.ExportCalendarFileParams) error {
	return s.calendarHandler.ExportCalendarFile(ctx, params)
}
func (s *Server) GetCalendarFeed(ctx echo.Context, params generated.GetCalendarFeedParams) error {
	return s.calendarHandler.GetCalendarFeed(ctx, params)
}
func (s *Server) ListLessonSeries(ctx echo.Context) error {
	return s.seriesHandler.ListLessonSeries(ctx)
}
func (s *Server) CreateLessonSeries(ctx echo.Context) error {
	return s.seriesHandler.CreateLessonSeries(ctx)
}
func (s *Server) GetLessonSeries(ctx echo.Context, id openapi_types.UUID) error {
	return s.seriesHandler.GetLessonSeries(ctx, id)
}
func (s *Server) UpdateLessonSeries(ctx echo.Context, id openapi_types.UUID) error {
	return s.seriesHandler.UpdateLessonSeries(ctx, id)
}
func (s *Server) DeleteLessonSeries(ctx echo.Context, id openapi_types.UUID) error {
	return s.seriesHandler.DeleteLessonSeries(ctx, id)
}
func (s *Server) ImportCalendarFile(ctx echo.Context) error {
	return s.calendarHandler.ImportCalendarFile(ctx)
}

// Finance & Export delegates
func (s *Server) GetFinanceSummary(ctx echo.Context, params generated.GetFinanceSummaryParams) error {
	return s.financeHandler.GetFinanceSummary(ctx, params)
}
func (s *Server) ListPayments(ctx echo.Context, params generated.ListPaymentsParams) error {
	return s.financeHandler.ListPayments(ctx, params)
}
func (s *Server) CreatePayment(ctx echo.Context) error {
	return s.financeHandler.CreatePayment(ctx)
}
func (s *Server) ListPartnerSettlements(ctx echo.Context, params generated.ListPartnerSettlementsParams) error {
	return s.financeHandler.ListPartnerSettlements(ctx, params)
}
func (s *Server) CreatePartnerPayout(ctx echo.Context) error {
	return s.financeHandler.CreatePartnerPayout(ctx)
}
func (s *Server) ExportClients(ctx echo.Context) error {
	return s.exportHandler.ExportClients(ctx)
}
func (s *Server) ExportLessons(ctx echo.Context, params generated.ExportLessonsParams) error {
	return s.exportHandler.ExportLessons(ctx, params)
}
func (s *Server) ExportPayments(ctx echo.Context, params generated.ExportPaymentsParams) error {
	return s.exportHandler.ExportPayments(ctx, params)
}

// Analytics delegates
func (s *Server) GetAnalyticsOverview(ctx echo.Context, params generated.GetAnalyticsOverviewParams) error {
	return s.analyticsHandler.GetAnalyticsOverview(ctx, params)
}
func (s *Server) GetAnalyticsDynamics(ctx echo.Context, params generated.GetAnalyticsDynamicsParams) error {
	return s.analyticsHandler.GetAnalyticsDynamics(ctx, params)
}
func (s *Server) GetAnalyticsFormats(ctx echo.Context, params generated.GetAnalyticsFormatsParams) error {
	return s.analyticsHandler.GetAnalyticsFormats(ctx, params)
}
func (s *Server) GetAnalyticsClients(ctx echo.Context, params generated.GetAnalyticsClientsParams) error {
	return s.analyticsHandler.GetAnalyticsClients(ctx, params)
}
func (s *Server) GetAnalyticsForecast(ctx echo.Context, params generated.GetAnalyticsForecastParams) error {
	return s.analyticsHandler.GetAnalyticsForecast(ctx, params)
}
func (s *Server) GetAnalyticsTags(ctx echo.Context, params generated.GetAnalyticsTagsParams) error {
	return s.analyticsHandler.GetAnalyticsTags(ctx, params)
}

// Dashboard delegates
func (s *Server) GetDashboardMetrics(ctx echo.Context, params generated.GetDashboardMetricsParams) error {
	return s.dashboardHandler.GetDashboardMetrics(ctx, params)
}
func (s *Server) GetDashboardSummary(ctx echo.Context) error {
	return s.dashboardHandler.GetDashboardSummary(ctx)
}

// Ensure interface compliance
var _ generated.ServerInterface = (*Server)(nil)

