package schedule

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/auth"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/response"
)

// SeriesHandler обрабатывает операции с регулярными сериями уроков.
type SeriesHandler struct {
	scheduleService *schedule.Service
	authHandler     *auth.Handler
}

func NewSeriesHandler(scheduleService *schedule.Service, authHandler *auth.Handler) *SeriesHandler {
	return &SeriesHandler{
		scheduleService: scheduleService,
		authHandler:     authHandler,
	}
}

func MapSeriesToResponse(s *domain.LessonSeries) generated.LessonSeriesResponse {
	var locPtr, notesPtr *string
	if s.LocationOrURL != "" {
		locPtr = &s.LocationOrURL
	}
	if s.Notes != "" {
		notesPtr = &s.Notes
	}
	var untilDatePtr *openapi_types.Date
	if s.UntilDate != nil {
		ud := openapi_types.Date{Time: *s.UntilDate}
		untilDatePtr = &ud
	}

	return generated.LessonSeriesResponse{
		Id:              s.ID,
		TeacherId:       s.TeacherID,
		ClientId:        s.ClientID,
		ClassroomId:     s.ClassroomID,
		Title:           s.Title,
		Rrule:           s.RRULE,
		StartTimeOfDay:  s.StartTimeOfDay,
		DurationMinutes: s.DurationMinutes,
		Format:          generated.LessonFormat(s.Format),
		LocationOrUrl:   locPtr,
		Notes:           notesPtr,
		StartDate:       openapi_types.Date{Time: s.StartDate},
		UntilDate:       untilDatePtr,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}

func (h *SeriesHandler) ListLessonSeries(c echo.Context) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	seriesList, err := h.scheduleService.ListSeries(c.Request().Context(), claims.UserID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list lesson series")
	}

	resp := make([]generated.LessonSeriesResponse, 0, len(seriesList))
	for _, s := range seriesList {
		resp = append(resp, MapSeriesToResponse(s))
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *SeriesHandler) CreateLessonSeries(c echo.Context) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	var req generated.CreateLessonSeriesRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	var untilDate *time.Time
	if req.UntilDate != nil {
		t := req.UntilDate.Time
		untilDate = &t
	}

	loc := ""
	if req.LocationOrUrl != nil {
		loc = *req.LocationOrUrl
	}
	notes := ""
	if req.Notes != nil {
		notes = *req.Notes
	}

	series, err := h.scheduleService.CreateSeries(c.Request().Context(), schedule.CreateSeriesInput{
		TeacherID:       claims.UserID,
		ClientID:        req.ClientId,
		ClassroomID:     req.ClassroomId,
		Title:           req.Title,
		RRULE:           req.Rrule,
		StartTimeOfDay:  req.StartTimeOfDay,
		DurationMinutes: req.DurationMinutes,
		Format:          domain.LessonFormat(req.Format),
		LocationOrURL:   loc,
		Notes:           notes,
		StartDate:       req.StartDate.Time,
		UntilDate:       untilDate,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidRRULE),
			errors.Is(err, domain.ErrInvalidSeriesDate),
			errors.Is(err, domain.ErrInvalidDuration),
			errors.Is(err, domain.ErrInvalidLessonFormat):
			return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		case errors.Is(err, domain.ErrUnauthorizedLessonAction):
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", err.Error())
		case errors.Is(err, domain.ErrClientNotFound):
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create lesson series")
		}
	}

	return c.JSON(http.StatusCreated, MapSeriesToResponse(series))
}

func (h *SeriesHandler) GetLessonSeries(c echo.Context, id openapi_types.UUID) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	series, err := h.scheduleService.GetSeries(c.Request().Context(), id, claims.UserID, claims.Role)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrLessonSeriesNotFound):
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		case errors.Is(err, domain.ErrUnauthorizedLessonAction):
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", err.Error())
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get lesson series")
		}
	}

	return c.JSON(http.StatusOK, MapSeriesToResponse(series))
}

func (h *SeriesHandler) UpdateLessonSeries(c echo.Context, id openapi_types.UUID) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	var req generated.UpdateLessonSeriesRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	var format *domain.LessonFormat
	if req.Format != nil {
		f := domain.LessonFormat(*req.Format)
		format = &f
	}

	var untilDate *time.Time
	if req.UntilDate != nil {
		t := req.UntilDate.Time
		untilDate = &t
	}

	series, err := h.scheduleService.UpdateSeries(c.Request().Context(), schedule.UpdateSeriesInput{
		SeriesID:        id,
		CallerID:        claims.UserID,
		CallerRole:      claims.Role,
		ClientID:        req.ClientId,
		ClassroomID:     req.ClassroomId,
		Title:           req.Title,
		RRULE:           req.Rrule,
		StartTimeOfDay:  req.StartTimeOfDay,
		DurationMinutes: req.DurationMinutes,
		Format:          format,
		LocationOrURL:   req.LocationOrUrl,
		Notes:           req.Notes,
		UntilDate:       untilDate,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrLessonSeriesNotFound):
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		case errors.Is(err, domain.ErrUnauthorizedLessonAction):
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", err.Error())
		case errors.Is(err, domain.ErrInvalidRRULE),
			errors.Is(err, domain.ErrInvalidSeriesDate),
			errors.Is(err, domain.ErrInvalidDuration),
			errors.Is(err, domain.ErrInvalidLessonFormat):
			return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update lesson series")
		}
	}

	return c.JSON(http.StatusOK, MapSeriesToResponse(series))
}

func (h *SeriesHandler) DeleteLessonSeries(c echo.Context, id openapi_types.UUID) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	err := h.scheduleService.DeleteSeries(c.Request().Context(), id, claims.UserID, claims.Role)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrLessonSeriesNotFound):
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		case errors.Is(err, domain.ErrUnauthorizedLessonAction):
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", err.Error())
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete lesson series")
		}
	}

	return c.NoContent(http.StatusNoContent)
}

