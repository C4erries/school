package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver/generated"
)

// ListTags реализует GET /tags.
func (h *APIHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	tags, err := h.crmService.ListTags(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list tags")
		return
	}

	resp := make([]generated.TagResponse, 0, len(tags))
	for _, t := range tags {
		resp = append(resp, generated.TagResponse{
			Id:            t.ID,
			TeacherId:     t.TeacherID,
			Name:          t.Name,
			SchoolPercent: t.SchoolPercent,
			Color:         t.Color,
			CreatedAt:     t.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// CreateTag реализует POST /tags.
func (h *APIHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can create tags")
		return
	}

	var req generated.CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	percent := 0
	if req.SchoolPercent != nil {
		percent = *req.SchoolPercent
	}

	color := "indigo"
	if req.Color != nil && *req.Color != "" {
		color = *req.Color
	}

	tag, err := h.crmService.CreateTag(r.Context(), claims.UserID, req.Name, percent, color)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTagName) || errors.Is(err, domain.ErrInvalidSchoolPercent) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create tag")
		return
	}

	writeJSON(w, http.StatusCreated, generated.TagResponse{
		Id:            tag.ID,
		TeacherId:     tag.TeacherID,
		Name:          tag.Name,
		SchoolPercent: tag.SchoolPercent,
		Color:         tag.Color,
		CreatedAt:     tag.CreatedAt,
	})
}

// DeleteTag реализует DELETE /tags/{id}.
func (h *APIHandler) DeleteTag(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	err := h.crmService.DeleteTag(r.Context(), id, claims.UserID, claims.Role)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTagNotFound):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "tag not found")
		case errors.Is(err, domain.ErrUnauthorizedTagAction):
			writeError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete tag")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
