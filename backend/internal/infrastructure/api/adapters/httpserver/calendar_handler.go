package httpserver

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver/generated"
)

// buildCalendarURLs формирует feed_url и webcal_url на основе текущего хоста запроса.
func buildCalendarURLs(r *http.Request, token uuid.UUID) (string, string) {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := r.Host
	if host == "" {
		host = "localhost:8080"
	}

	feedURL := fmt.Sprintf("%s://%s/api/v1/integrations/calendar/feed.ics?token=%s", scheme, host, token.String())
	webcalURL := fmt.Sprintf("webcal://%s/api/v1/integrations/calendar/feed.ics?token=%s", host, token.String())

	return feedURL, webcalURL
}

// GetCalendarSettings реализует GET /integrations/calendar/settings.
func (h *APIHandler) GetCalendarSettings(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view calendar settings")
		return
	}

	if h.calendarService == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "calendar service not configured")
		return
	}

	token, err := h.calendarService.GetSettings(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve calendar token")
		return
	}

	feedURL, webcalURL := buildCalendarURLs(r, token)

	resp := generated.CalendarSettingsResponse{
		CalendarToken: token,
		FeedUrl:       feedURL,
		WebcalUrl:     webcalURL,
	}

	writeJSON(w, http.StatusOK, resp)
}

// RotateCalendarToken реализует POST /integrations/calendar/rotate-token.
func (h *APIHandler) RotateCalendarToken(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can rotate calendar token")
		return
	}

	if h.calendarService == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "calendar service not configured")
		return
	}

	newToken, err := h.calendarService.RotateToken(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to rotate calendar token")
		return
	}

	feedURL, webcalURL := buildCalendarURLs(r, newToken)

	resp := generated.CalendarSettingsResponse{
		CalendarToken: newToken,
		FeedUrl:       feedURL,
		WebcalUrl:     webcalURL,
	}

	writeJSON(w, http.StatusOK, resp)
}

// ExportCalendarFile реализует GET /integrations/calendar/export.
func (h *APIHandler) ExportCalendarFile(w http.ResponseWriter, r *http.Request, params generated.ExportCalendarFileParams) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can export calendar")
		return
	}

	if h.calendarService == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "calendar service not configured")
		return
	}

	data, err := h.calendarService.ExportCalendarICS(r.Context(), claims.UserID, params.From, params.To)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to export calendar")
		return
	}

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"schedule.ics\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// GetCalendarFeed реализует публичный эндпоинт GET /integrations/calendar/feed.ics.
// Авторизация выполняется исключительно по query-параметру token.
func (h *APIHandler) GetCalendarFeed(w http.ResponseWriter, r *http.Request, params generated.GetCalendarFeedParams) {
	if h.calendarService == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "calendar service not configured")
		return
	}

	data, err := h.calendarService.GetFeedICS(r.Context(), params.Token)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or revoked calendar token")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to generate calendar feed")
		return
	}

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "inline; filename=\"feed.ics\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
