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
