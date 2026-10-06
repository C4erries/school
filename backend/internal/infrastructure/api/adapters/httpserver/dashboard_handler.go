package httpserver

import (
	"net/http"

	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver/generated"
)

// GetDashboardMetrics реализует GET /dashboard/metrics.
func (h *APIHandler) GetDashboardMetrics(w http.ResponseWriter, r *http.Request, params generated.GetDashboardMetricsParams) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view dashboard metrics")
		return
	}

	metrics, err := h.dashboardService.GetMetrics(r.Context(), claims.UserID, params.From, params.To)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate dashboard metrics")
		return
	}

	resp := generated.DashboardMetricsResponse{
		GrossPotentialRevenue: float32(metrics.GrossPotentialRevenue),
		NetIncome:             float32(metrics.NetIncome),
		AverageRate:           float32(metrics.AverageRate),
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetDashboardSummary реализует GET /dashboard/summary.
func (h *APIHandler) GetDashboardSummary(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view dashboard summary")
		return
	}

	summary, err := h.dashboardService.GetSummary(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate dashboard summary")
		return
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

	resp := generated.DashboardSummaryResponse{
		TodayLessons:      todayLessons,
		FinancialSnapshot: fin,
	}

	writeJSON(w, http.StatusOK, resp)
}
