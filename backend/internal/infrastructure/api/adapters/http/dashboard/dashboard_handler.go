package dashboard

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/C4erries/school/backend/internal/application/dashboard"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/auth"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/response"
)

// DashboardHandler обрабатывает агрегированные метрики для главной страницы.
type DashboardHandler struct {
	dashboardService *dashboard.Service
	authHandler      *auth.Handler
}

func NewDashboardHandler(dashboardService *dashboard.Service, authHandler *auth.Handler) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
		authHandler:      authHandler,
	}
}

func (h *DashboardHandler) GetDashboardMetrics(c echo.Context, params generated.GetDashboardMetricsParams) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view dashboard metrics")
	}

	metrics, err := h.dashboardService.GetMetrics(c.Request().Context(), claims.UserID, params.From, params.To)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate dashboard metrics")
	}

	return c.JSON(http.StatusOK, generated.DashboardMetricsResponse{
		GrossPotentialRevenue: float32(metrics.GrossPotentialRevenue),
		NetIncome:             float32(metrics.NetIncome),
		AverageRate:           float32(metrics.AverageRate),
	})
}

func (h *DashboardHandler) GetDashboardSummary(c echo.Context) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view dashboard summary")
	}

	summary, err := h.dashboardService.GetSummary(c.Request().Context(), claims.UserID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate dashboard summary")
	}

	todayLessons := make([]generated.TodayLesson, 0, len(summary.TodayLessons))
	for _, l := range summary.TodayLessons {
		todayLessons = append(todayLessons, generated.TodayLesson{
			Id:             l.ID,
			ClientId:       l.ClientID,
			ClientName:     l.ClientName,
			StartAt:        l.StartAt,
			EndAt:          l.EndAt,
			Format:         generated.LessonFormat(l.Format),
			Status:         generated.LessonStatus(l.Status),
			LocationType:   l.LocationType,
			OnlineLink:     l.OnlineLink,
			ClassroomName:  l.ClassroomName,
			ClassroomColor: l.ClassroomColor,
		})
	}

	fin := generated.FinancialSnapshot{
		MonthEarned:        float32(summary.FinancialSnapshot.MonthEarned),
		MonthForecast:      float32(summary.FinancialSnapshot.MonthForecast),
		TotalDebts:         float32(summary.FinancialSnapshot.TotalDebts),
		ActiveClientsCount: summary.FinancialSnapshot.ActiveClientsCount,
		WeeklyHours:        float32(summary.FinancialSnapshot.WeeklyHours),
	}

	return c.JSON(http.StatusOK, generated.DashboardSummaryResponse{
		TodayLessons:      todayLessons,
		FinancialSnapshot: fin,
	})
}

