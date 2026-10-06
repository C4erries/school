package httpserver

import (
	"net/http"

	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver/generated"
)

// GetAnalyticsOverview реализует GET /analytics/overview.
func (h *APIHandler) GetAnalyticsOverview(w http.ResponseWriter, r *http.Request, params generated.GetAnalyticsOverviewParams) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view analytics")
		return
	}

	if h.analyticsService == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "analytics service not configured")
		return
	}

	res, err := h.analyticsService.GetOverview(r.Context(), claims.UserID, params.From, params.To)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate analytics overview")
		return
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

	writeJSON(w, http.StatusOK, resp)
}

// GetAnalyticsDynamics реализует GET /analytics/dynamics.
func (h *APIHandler) GetAnalyticsDynamics(w http.ResponseWriter, r *http.Request, params generated.GetAnalyticsDynamicsParams) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view analytics")
		return
	}

	if h.analyticsService == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "analytics service not configured")
		return
	}

	intervalStr := "month"
	if params.Interval != nil {
		intervalStr = string(*params.Interval)
	}

	points, err := h.analyticsService.GetDynamics(r.Context(), claims.UserID, intervalStr, params.From, params.To)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate analytics dynamics")
		return
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

	writeJSON(w, http.StatusOK, resp)
}

// GetAnalyticsFormats реализует GET /analytics/formats.
func (h *APIHandler) GetAnalyticsFormats(w http.ResponseWriter, r *http.Request, params generated.GetAnalyticsFormatsParams) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view analytics")
		return
	}

	if h.analyticsService == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "analytics service not configured")
		return
	}

	formats, err := h.analyticsService.GetFormats(r.Context(), claims.UserID, params.From, params.To)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate analytics formats")
		return
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

	writeJSON(w, http.StatusOK, resp)
}

// GetAnalyticsClients реализует GET /analytics/clients.
func (h *APIHandler) GetAnalyticsClients(w http.ResponseWriter, r *http.Request, params generated.GetAnalyticsClientsParams) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view analytics")
		return
	}

	if h.analyticsService == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "analytics service not configured")
		return
	}

	sortField := "hours"
	if params.Sort != nil {
		sortField = string(*params.Sort)
	}

	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}

	clientStats, err := h.analyticsService.GetClients(r.Context(), claims.UserID, sortField, limit, params.From, params.To)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate analytics clients")
		return
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

	writeJSON(w, http.StatusOK, resp)
}

// GetAnalyticsForecast реализует GET /analytics/forecast.
func (h *APIHandler) GetAnalyticsForecast(w http.ResponseWriter, r *http.Request, params generated.GetAnalyticsForecastParams) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view analytics forecast")
		return
	}

	if h.analyticsService == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "analytics service not configured")
		return
	}

	forecast, err := h.analyticsService.GetForecast(r.Context(), claims.UserID, params.From, params.To)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate analytics forecast")
		return
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

	writeJSON(w, http.StatusOK, resp)
}

// GetAnalyticsTags реализует GET /analytics/tags.
func (h *APIHandler) GetAnalyticsTags(w http.ResponseWriter, r *http.Request, params generated.GetAnalyticsTagsParams) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view analytics tags")
		return
	}

	if h.analyticsService == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "analytics service not configured")
		return
	}

	tags, err := h.analyticsService.GetTagStats(r.Context(), claims.UserID, params.From, params.To)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate analytics tags")
		return
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

	writeJSON(w, http.StatusOK, resp)
}

