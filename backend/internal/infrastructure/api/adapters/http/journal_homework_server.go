package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/C4erries/school/backend/internal/application/journal"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/response"
)

// MapLessonJournalToResponse преобразует доменную сущность LessonJournal в DTO ответа.
func MapLessonJournalToResponse(j *domain.LessonJournal) generated.LessonJournalResponse {
	var notesPtr *string
	if j.Notes != "" {
		notesPtr = &j.Notes
	}
	return generated.LessonJournalResponse{
		Id:               j.ID,
		LessonId:         j.LessonID,
		ClientId:         j.ClientID,
		TeacherId:        j.TeacherID,
		Topic:            j.Topic,
		Notes:            notesPtr,
		PerformanceScore: j.PerformanceScore,
		CreatedAt:        j.CreatedAt,
		UpdatedAt:        j.UpdatedAt,
	}
}

// MapHomeworkToResponse преобразует доменную сущность HomeworkAssignment в DTO ответа.
func MapHomeworkToResponse(hw *domain.HomeworkAssignment) generated.HomeworkAssignmentResponse {
	var descPtr, reviewPtr *string
	if hw.Description != "" {
		descPtr = &hw.Description
	}
	if hw.ReviewNotes != "" {
		reviewPtr = &hw.ReviewNotes
	}
	var dueDatePtr *openapi_types.Date
	if hw.DueDate != nil {
		d := openapi_types.Date{Time: *hw.DueDate}
		dueDatePtr = &d
	}
	return generated.HomeworkAssignmentResponse{
		Id:               hw.ID,
		ClientId:         hw.ClientID,
		TeacherId:        hw.TeacherID,
		AssignedLessonId: hw.AssignedLessonID,
		Title:            hw.Title,
		Description:      descPtr,
		DueDate:          dueDatePtr,
		Status:           generated.HomeworkAssignmentResponseStatus(hw.Status),
		ReviewNotes:      reviewPtr,
		CreatedAt:        hw.CreatedAt,
		UpdatedAt:        hw.UpdatedAt,
	}
}

// MapBundleToResponse преобразует бандл журнала и домашних заданий в DTO ответа.
func MapBundleToResponse(bundle *journal.LessonJournalBundle) generated.LessonJournalBundleResponse {
	var journalPtr *generated.LessonJournalResponse
	if bundle.Journal != nil {
		resp := MapLessonJournalToResponse(bundle.Journal)
		journalPtr = &resp
	}
	assigned := make([]generated.HomeworkAssignmentResponse, 0, len(bundle.AssignedHomeworks))
	for _, hw := range bundle.AssignedHomeworks {
		assigned = append(assigned, MapHomeworkToResponse(hw))
	}
	due := make([]generated.HomeworkAssignmentResponse, 0, len(bundle.DueHomeworks))
	for _, hw := range bundle.DueHomeworks {
		due = append(due, MapHomeworkToResponse(hw))
	}
	return generated.LessonJournalBundleResponse{
		Journal:           journalPtr,
		AssignedHomeworks: assigned,
		DueHomeworks:      due,
	}
}

// GetLessonJournal возвращает отчет по уроку и связанные домашние задания.
func (s *Server) GetLessonJournal(ctx echo.Context, id openapi_types.UUID) error {
	claims, ok := s.authHandler.Authenticate(ctx)
	if !ok {
		return nil
	}

	if s.journalService == nil {
		return response.Error(ctx, http.StatusNotImplemented, "NOT_IMPLEMENTED", "journal service not configured")
	}

	bundle, err := s.journalService.GetLessonJournalBundle(ctx.Request().Context(), id, claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrLessonNotFound) {
			return response.Error(ctx, http.StatusNotFound, "NOT_FOUND", "lesson not found")
		}
		if errors.Is(err, domain.ErrUnauthorizedLessonAction) || errors.Is(err, domain.ErrForbidden) {
			return response.Error(ctx, http.StatusForbidden, "FORBIDDEN", "access denied")
		}
		return response.Error(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get lesson journal")
	}

	return ctx.JSON(http.StatusOK, MapBundleToResponse(bundle))
}

// UpsertLessonJournal сохраняет отчет по проведенному уроку.
func (s *Server) UpsertLessonJournal(ctx echo.Context, id openapi_types.UUID) error {
	claims, ok := s.authHandler.Authenticate(ctx)
	if !ok {
		return nil
	}

	if s.journalService == nil {
		return response.Error(ctx, http.StatusNotImplemented, "NOT_IMPLEMENTED", "journal service not configured")
	}

	var req generated.UpsertLessonJournalRequest
	if err := ctx.Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
	}

	notes := ""
	if req.Notes != nil {
		notes = *req.Notes
	}

	j, err := s.journalService.UpsertLessonJournal(ctx.Request().Context(), claims.UserID, id, journal.UpsertJournalInput{
		Topic:            req.Topic,
		Notes:            notes,
		PerformanceScore: req.PerformanceScore,
	})
	if err != nil {
		if errors.Is(err, domain.ErrLessonNotFound) {
			return response.Error(ctx, http.StatusNotFound, "NOT_FOUND", "lesson not found")
		}
		if errors.Is(err, domain.ErrUnauthorizedLessonAction) || errors.Is(err, domain.ErrForbidden) {
			return response.Error(ctx, http.StatusForbidden, "FORBIDDEN", "access denied")
		}
		if errors.Is(err, domain.ErrEmptyTopic) || errors.Is(err, domain.ErrInvalidPerformanceScore) {
			return response.Error(ctx, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		}
		return response.Error(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to save lesson journal")
	}

	return ctx.JSON(http.StatusOK, MapLessonJournalToResponse(j))
}

// GetClientJournals возвращает список всех отчетов по занятиям ученика.
func (s *Server) GetClientJournals(ctx echo.Context, id openapi_types.UUID) error {
	claims, ok := s.authHandler.Authenticate(ctx)
	if !ok {
		return nil
	}

	if s.journalService == nil {
		return response.Error(ctx, http.StatusNotImplemented, "NOT_IMPLEMENTED", "journal service not configured")
	}

	journals, err := s.journalService.GetClientJournals(ctx.Request().Context(), id, claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrClientNotFound) {
			return response.Error(ctx, http.StatusNotFound, "NOT_FOUND", "client not found")
		}
		if errors.Is(err, domain.ErrForbidden) {
			return response.Error(ctx, http.StatusForbidden, "FORBIDDEN", "access denied")
		}
		return response.Error(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list client journals")
	}

	resp := make([]generated.LessonJournalResponse, 0, len(journals))
	for _, j := range journals {
		resp = append(resp, MapLessonJournalToResponse(j))
	}

	return ctx.JSON(http.StatusOK, resp)
}

// ListClientHomework возвращает список всех домашних заданий ученика.
func (s *Server) ListClientHomework(ctx echo.Context, id openapi_types.UUID, params generated.ListClientHomeworkParams) error {
	claims, ok := s.authHandler.Authenticate(ctx)
	if !ok {
		return nil
	}

	if s.journalService == nil {
		return response.Error(ctx, http.StatusNotImplemented, "NOT_IMPLEMENTED", "journal service not configured")
	}

	var statusFilter *domain.HomeworkStatus
	if params.Status != nil {
		st := domain.HomeworkStatus(*params.Status)
		statusFilter = &st
	}

	hws, err := s.journalService.ListClientHomework(ctx.Request().Context(), id, claims.UserID, statusFilter)
	if err != nil {
		if errors.Is(err, domain.ErrClientNotFound) {
			return response.Error(ctx, http.StatusNotFound, "NOT_FOUND", "client not found")
		}
		if errors.Is(err, domain.ErrForbidden) {
			return response.Error(ctx, http.StatusForbidden, "FORBIDDEN", "access denied")
		}
		return response.Error(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list client homework")
	}

	resp := make([]generated.HomeworkAssignmentResponse, 0, len(hws))
	for _, hw := range hws {
		resp = append(resp, MapHomeworkToResponse(hw))
	}

	return ctx.JSON(http.StatusOK, resp)
}

// CreateHomework выдает новое домашнее задание ученику.
func (s *Server) CreateHomework(ctx echo.Context, id openapi_types.UUID) error {
	claims, ok := s.authHandler.Authenticate(ctx)
	if !ok {
		return nil
	}

	if s.journalService == nil {
		return response.Error(ctx, http.StatusNotImplemented, "NOT_IMPLEMENTED", "journal service not configured")
	}

	var req generated.CreateHomeworkRequest
	if err := ctx.Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
	}

	desc := ""
	if req.Description != nil {
		desc = *req.Description
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		d := req.DueDate.Time
		dueDate = &d
	}

	hw, err := s.journalService.CreateHomework(ctx.Request().Context(), journal.CreateHomeworkInput{
		ClientID:         id,
		TeacherID:        claims.UserID,
		AssignedLessonID: req.AssignedLessonId,
		Title:            req.Title,
		Description:      desc,
		DueDate:          dueDate,
	})
	if err != nil {
		if errors.Is(err, domain.ErrClientNotFound) || errors.Is(err, domain.ErrLessonNotFound) {
			return response.Error(ctx, http.StatusNotFound, "NOT_FOUND", err.Error())
		}
		if errors.Is(err, domain.ErrForbidden) || errors.Is(err, domain.ErrUnauthorizedLessonAction) {
			return response.Error(ctx, http.StatusForbidden, "FORBIDDEN", "access denied")
		}
		if errors.Is(err, domain.ErrEmptyHomeworkTitle) {
			return response.Error(ctx, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		}
		return response.Error(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create homework")
	}

	return ctx.JSON(http.StatusCreated, MapHomeworkToResponse(hw))
}

// UpdateHomeworkStatus обновляет статус выполнения и рецензию на задание.
func (s *Server) UpdateHomeworkStatus(ctx echo.Context, id openapi_types.UUID) error {
	claims, ok := s.authHandler.Authenticate(ctx)
	if !ok {
		return nil
	}

	if s.journalService == nil {
		return response.Error(ctx, http.StatusNotImplemented, "NOT_IMPLEMENTED", "journal service not configured")
	}

	var req generated.UpdateHomeworkStatusRequest
	if err := ctx.Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
	}

	reviewNotes := ""
	if req.ReviewNotes != nil {
		reviewNotes = *req.ReviewNotes
	}

	hw, err := s.journalService.UpdateHomeworkStatus(ctx.Request().Context(), journal.UpdateHomeworkStatusInput{
		ID:          id,
		TeacherID:   claims.UserID,
		Status:      domain.HomeworkStatus(req.Status),
		ReviewNotes: reviewNotes,
	})
	if err != nil {
		if errors.Is(err, domain.ErrHomeworkNotFound) {
			return response.Error(ctx, http.StatusNotFound, "NOT_FOUND", "homework not found")
		}
		if errors.Is(err, domain.ErrForbidden) {
			return response.Error(ctx, http.StatusForbidden, "FORBIDDEN", "access denied")
		}
		if errors.Is(err, domain.ErrInvalidHomeworkStatus) {
			return response.Error(ctx, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		}
		return response.Error(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update homework")
	}

	return ctx.JSON(http.StatusOK, MapHomeworkToResponse(hw))
}

// DeleteHomework удаляет домашнее задание.
func (s *Server) DeleteHomework(ctx echo.Context, id openapi_types.UUID) error {
	claims, ok := s.authHandler.Authenticate(ctx)
	if !ok {
		return nil
	}

	if s.journalService == nil {
		return response.Error(ctx, http.StatusNotImplemented, "NOT_IMPLEMENTED", "journal service not configured")
	}

	err := s.journalService.DeleteHomework(ctx.Request().Context(), id, claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrHomeworkNotFound) {
			return response.Error(ctx, http.StatusNotFound, "NOT_FOUND", "homework not found")
		}
		if errors.Is(err, domain.ErrForbidden) {
			return response.Error(ctx, http.StatusForbidden, "FORBIDDEN", "access denied")
		}
		return response.Error(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete homework")
	}

	return ctx.NoContent(http.StatusNoContent)
}
