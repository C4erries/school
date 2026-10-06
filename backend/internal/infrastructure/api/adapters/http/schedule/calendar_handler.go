package schedule

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/C4erries/school/backend/internal/application/calendar"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/auth"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/response"
)

// CalendarHandler обрабатывает интеграцию календаря iCalendar (.ics).
type CalendarHandler struct {
	calendarService *calendar.Service
	authHandler     *auth.Handler
}

func NewCalendarHandler(calendarService *calendar.Service, authHandler *auth.Handler) *CalendarHandler {
	return &CalendarHandler{
		calendarService: calendarService,
		authHandler:     authHandler,
	}
}

func buildCalendarURLs(c echo.Context, token uuid.UUID) (string, string) {
	r := c.Request()
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

func (h *CalendarHandler) GetCalendarSettings(c echo.Context) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view calendar settings")
	}

	if h.calendarService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "calendar service not configured")
	}

	token, err := h.calendarService.GetSettings(c.Request().Context(), claims.UserID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve calendar token")
	}

	feedURL, webcalURL := buildCalendarURLs(c, token)

	return c.JSON(http.StatusOK, generated.CalendarSettingsResponse{
		CalendarToken: token,
		FeedUrl:       feedURL,
		WebcalUrl:     webcalURL,
	})
}

func (h *CalendarHandler) RotateCalendarToken(c echo.Context) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can rotate calendar token")
	}

	if h.calendarService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "calendar service not configured")
	}

	newToken, err := h.calendarService.RotateToken(c.Request().Context(), claims.UserID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to rotate calendar token")
	}

	feedURL, webcalURL := buildCalendarURLs(c, newToken)

	return c.JSON(http.StatusOK, generated.CalendarSettingsResponse{
		CalendarToken: newToken,
		FeedUrl:       feedURL,
		WebcalUrl:     webcalURL,
	})
}

func (h *CalendarHandler) ExportCalendarFile(c echo.Context, params generated.ExportCalendarFileParams) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can export calendar")
	}

	if h.calendarService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "calendar service not configured")
	}

	data, err := h.calendarService.ExportCalendarICS(c.Request().Context(), claims.UserID, params.From, params.To)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to export calendar")
	}

	c.Response().Header().Set("Content-Type", "text/calendar; charset=utf-8")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=\"schedule.ics\"")
	return c.Blob(http.StatusOK, "text/calendar; charset=utf-8", data)
}

func (h *CalendarHandler) GetCalendarFeed(c echo.Context, params generated.GetCalendarFeedParams) error {
	if h.calendarService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "calendar service not configured")
	}

	data, err := h.calendarService.GetFeedICS(c.Request().Context(), params.Token)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or revoked calendar token")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to generate calendar feed")
	}

	c.Response().Header().Set("Content-Type", "text/calendar; charset=utf-8")
	c.Response().Header().Set("Content-Disposition", "inline; filename=\"feed.ics\"")
	return c.Blob(http.StatusOK, "text/calendar; charset=utf-8", data)
}

func (h *CalendarHandler) ImportCalendarFile(c echo.Context) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can import calendar")
	}

	if h.calendarService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "calendar service not configured")
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "file is required")
	}

	f, err := fileHeader.Open()
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "failed to open uploaded file")
	}
	defer f.Close()

	res, err := h.calendarService.ImportICS(c.Request().Context(), claims.UserID, f)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_CALENDAR_FILE", err.Error())
	}

	msg := "Календарь успешно импортирован"
	return c.JSON(http.StatusOK, generated.CalendarImportResponse{
		ImportedLessons: res.ImportedLessons,
		ImportedSeries:  res.ImportedSeries,
		SkippedEvents:   res.SkippedEvents,
		Message:         &msg,
	})
}


