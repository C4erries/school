package schedule

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/auth"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/response"
)

// LessonHandler обрабатывает операции с расписанием уроков.
type LessonHandler struct {
	scheduleService *schedule.Service
	authHandler     *auth.Handler
}

func NewLessonHandler(scheduleService *schedule.Service, authHandler *auth.Handler) *LessonHandler {
	return &LessonHandler{
		scheduleService: scheduleService,
		authHandler:     authHandler,
	}
}

func MapLessonToResponse(l *domain.Lesson) generated.LessonResponse {
	var locPtr, notesPtr, reasonPtr *string
	if l.LocationOrURL != "" {
		locPtr = &l.LocationOrURL
	}
	if l.Notes != "" {
		notesPtr = &l.Notes
	}
	if l.CancelReason != "" {
		reasonPtr = &l.CancelReason
	}

	title := l.Title
	if title == "" {
		title = "Занятие"
	}

	return generated.LessonResponse{
		Id:            l.ID,
		TeacherId:     l.TeacherID,
		ClientId:      l.ClientID,
		Title:         title,
		ClassroomId:   l.ClassroomID,
		StartTime:     l.StartTime,
		EndTime:       l.EndTime,
		Format:        generated.LessonFormat(l.Format),
		LocationOrUrl: locPtr,
		Status:        generated.LessonStatus(l.Status),
		Notes:         notesPtr,
		CancelReason:  reasonPtr,
		CreatedAt:     l.CreatedAt,
		UpdatedAt:     l.UpdatedAt,
	}
}

func (h *LessonHandler) ListLessons(c echo.Context, params generated.ListLessonsParams) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	filter := schedule.LessonFilter{
		ClassroomID: params.ClassroomId,
		From:        params.From,
		To:          params.To,
	}

	if params.Status != nil {
		st := domain.LessonStatus(*params.Status)
		filter.Status = &st
	}

	switch claims.Role {
	case domain.RoleTeacher:
		filter.TeacherID = &claims.UserID
		if params.ClientId != nil {
			filter.ClientID = params.ClientId
		}
	case domain.RoleOwner:
		filter.TeacherID = params.TeacherId
		filter.ClientID = params.ClientId
	default:
		filter.TeacherID = &claims.UserID
	}

	lessons, err := h.scheduleService.ListLessons(c.Request().Context(), filter)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list lessons")
	}

	resp := make([]generated.LessonResponse, 0, len(lessons))
	for _, l := range lessons {
		resp = append(resp, MapLessonToResponse(l))
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *LessonHandler) CreateLesson(c echo.Context) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can create lessons")
	}

	var req generated.CreateLessonRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	teacherID := claims.UserID

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "Занятие"
	}

	var loc, notes string
	if req.LocationOrUrl != nil {
		loc = *req.LocationOrUrl
	}
	if req.Notes != nil {
		notes = *req.Notes
	}

	lesson, err := h.scheduleService.ScheduleLesson(c.Request().Context(), schedule.ScheduleLessonInput{
		TeacherID:     teacherID,
		ClientID:      req.ClientId,
		Title:         title,
		ClassroomID:   req.ClassroomId,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		Format:        domain.LessonFormat(req.Format),
		LocationOrURL: loc,
		Notes:         notes,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrClassroomCollision):
			return response.Error(c, http.StatusConflict, "CLASSROOM_COLLISION", err.Error())
		case errors.Is(err, domain.ErrInvalidTimeRange),
			errors.Is(err, domain.ErrInvalidLessonFormat),
			errors.Is(err, domain.ErrClassroomRequiredForOffline),
			errors.Is(err, domain.ErrClassroomNotFound):
			return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		case errors.Is(err, domain.ErrUnauthorizedLessonAction):
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", err.Error())
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create lesson")
		}
	}

	return c.JSON(http.StatusCreated, MapLessonToResponse(lesson))
}

func (h *LessonHandler) CompleteLesson(c echo.Context, id openapi_types.UUID) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	lesson, err := h.scheduleService.CompleteLesson(c.Request().Context(), id, claims.UserID, claims.Role)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrLessonNotFound):
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		case errors.Is(err, domain.ErrUnauthorizedLessonAction):
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", err.Error())
		case errors.Is(err, domain.ErrInvalidLessonStatus):
			return response.Error(c, http.StatusBadRequest, "INVALID_STATUS", err.Error())
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to complete lesson")
		}
	}

	return c.JSON(http.StatusOK, MapLessonToResponse(lesson))
}

func (h *LessonHandler) CancelLesson(c echo.Context, id openapi_types.UUID) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	var reason string
	var req generated.CancelLessonRequest
	if err := c.Bind(&req); err == nil && req.Reason != nil {
		reason = *req.Reason
	}

	lesson, err := h.scheduleService.CancelLesson(c.Request().Context(), id, claims.UserID, claims.Role, reason)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrLessonNotFound):
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		case errors.Is(err, domain.ErrUnauthorizedLessonAction):
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", err.Error())
		case errors.Is(err, domain.ErrInvalidLessonStatus):
			return response.Error(c, http.StatusBadRequest, "INVALID_STATUS", err.Error())
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to cancel lesson")
		}
	}

	return c.JSON(http.StatusOK, MapLessonToResponse(lesson))
}

func (h *LessonHandler) UpdateLesson(c echo.Context, id openapi_types.UUID) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "failed to read request body")
	}

	var req generated.UpdateLessonRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	var rawMap map[string]json.RawMessage
	_ = json.Unmarshal(bodyBytes, &rawMap)

	var format *domain.LessonFormat
	if req.Format != nil {
		f := domain.LessonFormat(*req.Format)
		format = &f
	}

	clearClassroom := false
	if rawVal, exists := rawMap["classroom_id"]; exists {
		if string(rawVal) == "null" || (req.ClassroomId != nil && *req.ClassroomId == openapi_types.UUID(uuid.Nil)) {
			clearClassroom = true
		}
	} else if req.ClassroomId != nil && *req.ClassroomId == openapi_types.UUID(uuid.Nil) {
		clearClassroom = true
	}

	if req.LocationOrUrl != nil && strings.TrimSpace(*req.LocationOrUrl) != "" && (req.ClassroomId == nil || *req.ClassroomId == openapi_types.UUID(uuid.Nil)) {
		clearClassroom = true
	}

	var locOrURL *string = req.LocationOrUrl
	emptyStr := ""
	if (req.LocationOrUrl != nil && strings.TrimSpace(*req.LocationOrUrl) == "") ||
		(req.ClassroomId != nil && *req.ClassroomId != openapi_types.UUID(uuid.Nil) && (req.LocationOrUrl == nil || strings.TrimSpace(*req.LocationOrUrl) == "")) {
		locOrURL = &emptyStr
	}

	lesson, err := h.scheduleService.UpdateLesson(c.Request().Context(), schedule.UpdateLessonInput{
		LessonID:       id,
		CallerID:       claims.UserID,
		CallerRole:     claims.Role,
		Title:          req.Title,
		ClientID:       req.ClientId,
		ClassroomID:    req.ClassroomId,
		ClearClassroom: clearClassroom,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		Format:         format,
		LocationOrURL:  locOrURL,
		Notes:          req.Notes,
		CancelReason:   req.CancelReason,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrLessonNotFound):
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		case errors.Is(err, domain.ErrUnauthorizedLessonAction):
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", err.Error())
		case errors.Is(err, domain.ErrClassroomCollision):
			return response.Error(c, http.StatusConflict, "CLASSROOM_COLLISION", err.Error())
		case errors.Is(err, domain.ErrInvalidTimeRange),
			errors.Is(err, domain.ErrInvalidLessonFormat),
			errors.Is(err, domain.ErrClassroomNotFound):
			return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update lesson")
		}
	}

	return c.JSON(http.StatusOK, MapLessonToResponse(lesson))
}

