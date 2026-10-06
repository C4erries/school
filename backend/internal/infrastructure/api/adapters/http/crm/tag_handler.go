package crm

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/C4erries/school/backend/internal/application/crm"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/auth"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/response"
)

// TagHandler обрабатывает работу с тегами и привязками клиентов.
type TagHandler struct {
	crmService  *crm.Service
	authHandler *auth.Handler
}

func NewTagHandler(crmService *crm.Service, authHandler *auth.Handler) *TagHandler {
	return &TagHandler{
		crmService:  crmService,
		authHandler: authHandler,
	}
}

func (h *TagHandler) ListTags(c echo.Context) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	tags, err := h.crmService.ListTags(c.Request().Context(), claims.UserID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list tags")
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

	return c.JSON(http.StatusOK, resp)
}

func (h *TagHandler) CreateTag(c echo.Context) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can create tags")
	}

	var req generated.CreateTagRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	percent := 0
	if req.SchoolPercent != nil {
		percent = *req.SchoolPercent
	}

	color := "indigo"
	if req.Color != nil && *req.Color != "" {
		color = *req.Color
	}

	tag, err := h.crmService.CreateTag(c.Request().Context(), claims.UserID, req.Name, percent, color)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTagName) || errors.Is(err, domain.ErrInvalidSchoolPercent) {
			return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create tag")
	}

	return c.JSON(http.StatusCreated, generated.TagResponse{
		Id:            tag.ID,
		TeacherId:     tag.TeacherID,
		Name:          tag.Name,
		SchoolPercent: tag.SchoolPercent,
		Color:         tag.Color,
		CreatedAt:     tag.CreatedAt,
	})
}

func (h *TagHandler) DeleteTag(c echo.Context, id openapi_types.UUID) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	err := h.crmService.DeleteTag(c.Request().Context(), id, claims.UserID, claims.Role)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTagNotFound):
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", "tag not found")
		case errors.Is(err, domain.ErrUnauthorizedTagAction):
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete tag")
		}
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *TagHandler) AssignClientTag(c echo.Context, id openapi_types.UUID) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	var req generated.AssignTagRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	client, err := h.crmService.AssignTagToClient(c.Request().Context(), id, req.TagId, claims.UserID, claims.Role)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrClientNotFound), errors.Is(err, domain.ErrTagNotFound):
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		case errors.Is(err, crm.ErrUnauthorizedAction), errors.Is(err, domain.ErrUnauthorizedTagAction):
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", err.Error())
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to assign tag")
		}
	}

	return c.JSON(http.StatusOK, MapClientToResponse(client))
}

func (h *TagHandler) RemoveClientTag(c echo.Context, id openapi_types.UUID, tagId openapi_types.UUID) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	client, err := h.crmService.RemoveTagFromClient(c.Request().Context(), id, tagId, claims.UserID, claims.Role)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrClientNotFound):
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		case errors.Is(err, crm.ErrUnauthorizedAction):
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", err.Error())
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to remove tag")
		}
	}

	return c.JSON(http.StatusOK, MapClientToResponse(client))
}

