package analytics

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/C4erries/school/backend/internal/application/analytics"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/auth"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/response"
)

// AnalyticsHandler обрабатывает эндпоинты аналитики и сводных отчетов.
type AnalyticsHandler struct {
	analyticsService *analytics.Service
	authHandler      *auth.Handler
}

func NewAnalyticsHandler(analyticsService *analytics.Service, authHandler *auth.Handler) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		authHandler:      authHandler,
	}
}

func (h *AnalyticsHandler) GetAnalyticsOverview(c echo.Context, params generated.GetAnalyticsOverviewParams) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view analytics")
	}

	if h.analyticsService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "analytics service not configured")
	}

	res, err := h.analyticsService.GetOverview(c.Request().Context(), claims.UserID, params.From, params.To)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate analytics overview")
	}

	resp := generated.AnalyticsOverviewResponse{
		From:                res.From,
		To:                  res.To,
		TotalLessons:        res.TotalLessons,
		CompletedLessons:    res.CompletedLessons,
		CancelledLessons:    res.CancelledLessons,
		CompletionRate:      res.CompletionRate,
		CompletedHours:      res.CompletedHours,
		GrossRevenue:        res.GrossRevenue,
		NetIncome:           res.NetIncome,
		EffectiveHourlyRate: res.EffectiveHourlyRate,
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AnalyticsHandler) GetAnalyticsDynamics(c echo.Context, params generated.GetAnalyticsDynamicsParams) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view analytics")
	}

	if h.analyticsService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "analytics service not configured")
	}

	intervalStr := "month"
	if params.Interval != nil {
		intervalStr = string(*params.Interval)
	}

	points, err := h.analyticsService.GetDynamics(c.Request().Context(), claims.UserID, intervalStr, params.From, params.To)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate analytics dynamics")
	}

	resp := make([]generated.AnalyticsDynamicsPoint, len(points))
	for i, p := range points {
		resp[i] = generated.AnalyticsDynamicsPoint{
			Label:          p.Label,
			From:           p.From,
			To:             p.To,
			CompletedHours: p.CompletedHours,
			NetIncome:      p.NetIncome,
			CompletedCount: p.CompletedCount,
			CancelledCount: p.CancelledCount,
		}
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AnalyticsHandler) GetAnalyticsFormats(c echo.Context, params generated.GetAnalyticsFormatsParams) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view analytics")
	}

	if h.analyticsService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "analytics service not configured")
	}

	formats, err := h.analyticsService.GetFormats(c.Request().Context(), claims.UserID, params.From, params.To)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate analytics formats")
	}

	resp := make([]generated.AnalyticsFormatStat, len(formats))
	for i, f := range formats {
		resp[i] = generated.AnalyticsFormatStat{
			Format:              generated.LessonFormat(f.Format),
			CompletedHours:      f.CompletedHours,
			NetIncome:           f.NetIncome,
			LessonsCount:        f.LessonsCount,
			HoursSharePercent:   f.HoursSharePercent,
			RevenueSharePercent: f.RevenueSharePercent,
		}
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AnalyticsHandler) GetAnalyticsClients(c echo.Context, params generated.GetAnalyticsClientsParams) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view analytics")
	}

	if h.analyticsService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "analytics service not configured")
	}

	sortField := "hours"
	if params.Sort != nil {
		sortField = string(*params.Sort)
	}

	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}

	clientStats, err := h.analyticsService.GetClients(c.Request().Context(), claims.UserID, sortField, limit, params.From, params.To)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate analytics clients")
	}

	resp := make([]generated.AnalyticsClientStat, len(clientStats))
	for i, cs := range clientStats {
		resp[i] = generated.AnalyticsClientStat{
			ClientId:       cs.ClientID,
			ClientName:     cs.ClientName,
			CompletedHours: cs.CompletedHours,
			NetIncome:      cs.NetIncome,
			CompletedCount: cs.CompletedCount,
			CancelledCount: cs.CancelledCount,
			AttendanceRate: cs.AttendanceRate,
		}
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AnalyticsHandler) GetAnalyticsForecast(c echo.Context, params generated.GetAnalyticsForecastParams) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view analytics forecast")
	}

	if h.analyticsService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "analytics service not configured")
	}

	forecast, err := h.analyticsService.GetForecast(c.Request().Context(), claims.UserID, params.From, params.To)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate analytics forecast")
	}

	byFormat := make([]generated.FormatForecast, len(forecast.ByFormat))
	for i, f := range forecast.ByFormat {
		byFormat[i] = generated.FormatForecast{
			Format:  generated.LessonFormat(f.Format),
			Hours:   f.Hours,
			Revenue: f.Revenue,
		}
	}

	resp := generated.AnalyticsForecastResponse{
		From:                      forecast.From,
		To:                        forecast.To,
		ScheduledLessons:          forecast.ScheduledLessons,
		ScheduledHours:            forecast.ScheduledHours,
		GrossPotentialRevenue:     forecast.GrossPotentialRevenue,
		PartnerCommissionExpected: forecast.PartnerCommissionExpected,
		NetPotentialIncome:        forecast.NetPotentialIncome,
		ByFormat:                  byFormat,
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AnalyticsHandler) GetAnalyticsTags(c echo.Context, params generated.GetAnalyticsTagsParams) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view analytics tags")
	}

	if h.analyticsService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "analytics service not configured")
	}

	tags, err := h.analyticsService.GetTagStats(c.Request().Context(), claims.UserID, params.From, params.To)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate analytics tags")
	}

	resp := make([]generated.TagStatResponse, len(tags))
	for i, t := range tags {
		resp[i] = generated.TagStatResponse{
			TagId:          t.TagID,
			TagName:        t.TagName,
			TagColor:       t.TagColor,
			StudentsCount:  t.StudentsCount,
			CompletedHours: t.CompletedHours,
			GrossRevenue:   t.GrossRevenue,
			NetIncome:      t.NetIncome,
		}
	}

	return c.JSON(http.StatusOK, resp)
}

