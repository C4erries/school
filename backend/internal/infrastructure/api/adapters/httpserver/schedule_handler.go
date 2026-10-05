package httpserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver/generated"
)

// ListClassrooms реализует GET /classrooms.
func (h *APIHandler) ListClassrooms(w http.ResponseWriter, r *http.Request) {
	_, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	classrooms, err := h.scheduleService.ListClassrooms(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list classrooms")
		return
	}

	resp := make([]generated.ClassroomResponse, 0, len(classrooms))
	for _, c := range classrooms {
		var desc *string
		if c.Description != "" {
			desc = &c.Description
		}
		resp = append(resp, generated.ClassroomResponse{
			Id:          c.ID,
			Name:        c.Name,
			Capacity:    c.Capacity,
			Color:       c.Color,
			Description: desc,
			CreatedAt:   c.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// CreateClassroom реализует POST /classrooms.
func (h *APIHandler) CreateClassroom(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only admin can create classrooms")
		return
	}

	var req generated.CreateClassroomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	var color, desc string
	if req.Color != nil {
		color = *req.Color
	}
	if req.Description != nil {
		desc = *req.Description
	}

	c, err := h.scheduleService.CreateClassroom(r.Context(), schedule.CreateClassroomInput{
		Name:        req.Name,
		Capacity:    req.Capacity,
		Color:       color,
		Description: desc,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidClassroomName) || errors.Is(err, domain.ErrInvalidClassroomCapacity) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create classroom")
		return
	}

	var descPtr *string
	if c.Description != "" {
		descPtr = &c.Description
	}

	writeJSON(w, http.StatusCreated, generated.ClassroomResponse{
		Id:          c.ID,
		Name:        c.Name,
		Capacity:    c.Capacity,
		Color:       c.Color,
		Description: descPtr,
		CreatedAt:   c.CreatedAt,
	})
}

// ListLessons реализует GET /lessons.
func (h *APIHandler) ListLessons(w http.ResponseWriter, r *http.Request, params generated.ListLessonsParams) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
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

	lessons, err := h.scheduleService.ListLessons(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list lessons")
		return
	}

	resp := make([]generated.LessonResponse, 0, len(lessons))
	for _, l := range lessons {
		resp = append(resp, mapLessonToResponse(l))
	}

	writeJSON(w, http.StatusOK, resp)
}

// CreateLesson реализует POST /lessons.
func (h *APIHandler) CreateLesson(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can create lessons")
		return
	}

	var req generated.CreateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
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

	lesson, err := h.scheduleService.ScheduleLesson(r.Context(), schedule.ScheduleLessonInput{
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
			writeError(w, http.StatusConflict, "CLASSROOM_COLLISION", err.Error())
		case errors.Is(err, domain.ErrInvalidTimeRange),
			errors.Is(err, domain.ErrInvalidLessonFormat),
			errors.Is(err, domain.ErrClassroomRequiredForOffline),
			errors.Is(err, domain.ErrClassroomNotFound):
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		case errors.Is(err, domain.ErrUnauthorizedLessonAction):
			writeError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create lesson")
		}
		return
	}

	writeJSON(w, http.StatusCreated, mapLessonToResponse(lesson))
}

// CompleteLesson реализует POST /lessons/{id}/complete.
func (h *APIHandler) CompleteLesson(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	lesson, err := h.scheduleService.CompleteLesson(r.Context(), id, claims.UserID, claims.Role)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrLessonNotFound):
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		case errors.Is(err, domain.ErrUnauthorizedLessonAction):
			writeError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
		case errors.Is(err, domain.ErrInvalidLessonStatus):
			writeError(w, http.StatusBadRequest, "INVALID_STATUS", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to complete lesson")
		}
		return
	}

	writeJSON(w, http.StatusOK, mapLessonToResponse(lesson))
}

// CancelLesson реализует POST /lessons/{id}/cancel.
func (h *APIHandler) CancelLesson(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	var reason string
	var req generated.CancelLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.Reason != nil {
		reason = *req.Reason
	}

	lesson, err := h.scheduleService.CancelLesson(r.Context(), id, claims.UserID, claims.Role, reason)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrLessonNotFound):
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		case errors.Is(err, domain.ErrUnauthorizedLessonAction):
			writeError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
		case errors.Is(err, domain.ErrInvalidLessonStatus):
			writeError(w, http.StatusBadRequest, "INVALID_STATUS", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to cancel lesson")
		}
		return
	}

	writeJSON(w, http.StatusOK, mapLessonToResponse(lesson))
}

// UpdateLesson реализует PATCH /lessons/{id}.
func (h *APIHandler) UpdateLesson(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "failed to read request body")
		return
	}

	var req generated.UpdateLessonRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
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

	// Если передан location_or_url без кабинета, гарантируем сброс classroom_id при переводе урока в онлайн
	if req.LocationOrUrl != nil && strings.TrimSpace(*req.LocationOrUrl) != "" && (req.ClassroomId == nil || *req.ClassroomId == openapi_types.UUID(uuid.Nil)) {
		clearClassroom = true
	}

	// Фикс сброса онлайна в оффлайн:
	// Если req.LocationOrUrl != nil и *req.LocationOrUrl == "" (или если передан валидный classroom_id и req.LocationOrUrl не передан/пустой), очищай lesson.LocationOrURL = "".
	// Таким образом старый онлайн URL гарантированно стирается из БД при переводе в оффлайн!
	var locOrURL *string = req.LocationOrUrl
	emptyStr := ""
	if (req.LocationOrUrl != nil && strings.TrimSpace(*req.LocationOrUrl) == "") ||
		(req.ClassroomId != nil && *req.ClassroomId != openapi_types.UUID(uuid.Nil) && (req.LocationOrUrl == nil || strings.TrimSpace(*req.LocationOrUrl) == "")) {
		locOrURL = &emptyStr
	}

	lesson, err := h.scheduleService.UpdateLesson(r.Context(), schedule.UpdateLessonInput{
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
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		case errors.Is(err, domain.ErrUnauthorizedLessonAction):
			writeError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
		case errors.Is(err, domain.ErrClassroomCollision):
			writeError(w, http.StatusConflict, "CLASSROOM_COLLISION", err.Error())
		case errors.Is(err, domain.ErrInvalidTimeRange),
			errors.Is(err, domain.ErrInvalidLessonFormat),
			errors.Is(err, domain.ErrClassroomNotFound):
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update lesson")
		}
		return
	}

	writeJSON(w, http.StatusOK, mapLessonToResponse(lesson))
}

func mapLessonToResponse(l *domain.Lesson) generated.LessonResponse {
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
