package crm

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/C4erries/school/backend/internal/application/crm"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/auth"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/response"
)

// ClientHandler обрабатывает CRUD операции над клиентами.
type ClientHandler struct {
	crmService  *crm.Service
	authHandler *auth.Handler
}

func NewClientHandler(crmService *crm.Service, authHandler *auth.Handler) *ClientHandler {
	return &ClientHandler{
		crmService:  crmService,
		authHandler: authHandler,
	}
}

func MapClientToResponse(c *domain.Client) generated.ClientResponse {
	rateIndiv := float32(c.RateIndividual)
	if rateIndiv == 0 && c.BaseRate > 0 {
		rateIndiv = float32(c.BaseRate)
	}

	var ratePair *float32
	if c.RatePair != nil {
		v := float32(*c.RatePair)
		ratePair = &v
	}

	var rateGroup *float32
	if c.RateGroup != nil {
		v := float32(*c.RateGroup)
		rateGroup = &v
	}

	baseRate := float32(c.BaseRate)
	if baseRate == 0 {
		baseRate = rateIndiv
	}

	tags := make([]generated.TagResponse, 0, len(c.Tags))
	for _, t := range c.Tags {
		tags = append(tags, generated.TagResponse{
			Id:            t.ID,
			TeacherId:     t.TeacherID,
			Name:          t.Name,
			SchoolPercent: t.SchoolPercent,
			Color:         t.Color,
			CreatedAt:     t.CreatedAt,
		})
	}

	balances := generated.ClientBalances{
		IndividualHours: float32(c.Balances.IndividualHours),
		PairHours:       float32(c.Balances.PairHours),
		GroupHours:      float32(c.Balances.GroupHours),
		TotalHours:      float32(c.Balances.TotalHours),
	}

	schoolPercentTag := c.SchoolPercentTag

	return generated.ClientResponse{
		Id:               c.ID,
		TeacherId:        c.TeacherID,
		Name:             c.Name,
		Phone:            c.Phone,
		BaseRate:         &baseRate,
		RateIndividual:   rateIndiv,
		RatePair:         ratePair,
		RateGroup:        rateGroup,
		SchoolPercentTag: &schoolPercentTag,
		IsArchived:       c.IsArchived,
		Tags:             tags,
		Balances:         balances,
		LastLessonAt:     c.LastLessonAt,
		CreatedAt:        c.CreatedAt,
	}
}

func (h *ClientHandler) ListClients(c echo.Context, params generated.ListClientsParams) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can list clients")
	}

	var filter []crm.ClientFilter
	if params.IsArchived != nil || (params.Search != nil && strings.TrimSpace(*params.Search) != "") {
		filter = append(filter, crm.ClientFilter{
			IsArchived: params.IsArchived,
			Search:     params.Search,
		})
	}

	clients, err := h.crmService.ListClients(c.Request().Context(), claims.UserID, filter...)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list clients")
	}

	resp := make([]generated.ClientResponse, 0, len(clients))
	for _, client := range clients {
		resp = append(resp, MapClientToResponse(client))
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *ClientHandler) CreateClient(c echo.Context) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can create clients")
	}

	var req generated.CreateClientRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "client name is required")
	}

	var rateIndividual float64
	if req.RateIndividual != nil && *req.RateIndividual > 0 {
		rateIndividual = float64(*req.RateIndividual)
	} else if req.BaseRate != nil && *req.BaseRate > 0 {
		rateIndividual = float64(*req.BaseRate)
	} else {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "rate_individual must be greater than 0")
	}

	var ratePair *float64
	if req.RatePair != nil {
		v := float64(*req.RatePair)
		ratePair = &v
	}

	var rateGroup *float64
	if req.RateGroup != nil {
		v := float64(*req.RateGroup)
		rateGroup = &v
	}

	schoolPercentTag := 0
	if req.SchoolPercentTag != nil {
		schoolPercentTag = *req.SchoolPercentTag
	}

	var tagIDs []openapi_types.UUID
	if req.TagIds != nil {
		tagIDs = *req.TagIds
	}

	isArchived := false
	if req.IsArchived != nil {
		isArchived = *req.IsArchived
	}

	client, err := h.crmService.CreateClientWithRates(c.Request().Context(), crm.CreateClientInput{
		TeacherID:        claims.UserID,
		Name:             name,
		Phone:            req.Phone,
		BaseRate:         rateIndividual,
		RateIndividual:   rateIndividual,
		RatePair:         ratePair,
		RateGroup:        rateGroup,
		SchoolPercentTag: schoolPercentTag,
		IsArchived:       isArchived,
		TagIDs:           tagIDs,
	})
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create client")
	}

	return c.JSON(http.StatusCreated, MapClientToResponse(client))
}

func (h *ClientHandler) GetClient(c echo.Context, id openapi_types.UUID) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	client, err := h.crmService.GetClient(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrClientNotFound) {
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", "client not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get client")
	}

	if claims.Role != domain.RoleOwner && client.TeacherID != claims.UserID {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
	}

	return c.JSON(http.StatusOK, MapClientToResponse(client))
}

func (h *ClientHandler) UpdateClient(c echo.Context, id openapi_types.UUID) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	var req generated.UpdateClientRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	var rateIndiv, ratePair, rateGroup *float64
	if req.RateIndividual != nil {
		v := float64(*req.RateIndividual)
		if v < 0 {
			return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "rate_individual cannot be negative")
		}
		rateIndiv = &v
	}
	if req.RatePair != nil {
		v := float64(*req.RatePair)
		if v < 0 {
			return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "rate_pair cannot be negative")
		}
		ratePair = &v
	}
	if req.RateGroup != nil {
		v := float64(*req.RateGroup)
		if v < 0 {
			return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "rate_group cannot be negative")
		}
		rateGroup = &v
	}

	client, err := h.crmService.UpdateClient(c.Request().Context(), crm.UpdateClientInput{
		ID:             id,
		CallerID:       claims.UserID,
		CallerRole:     claims.Role,
		Name:           req.Name,
		Phone:          req.Phone,
		RateIndividual: rateIndiv,
		RatePair:       ratePair,
		RateGroup:      rateGroup,
		IsArchived:     req.IsArchived,
		TagIDs:         req.TagIds,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrClientNotFound):
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", "client not found")
		case errors.Is(err, crm.ErrUnauthorizedAction):
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update client")
		}
	}

	return c.JSON(http.StatusOK, MapClientToResponse(client))
}

func (h *ClientHandler) DeleteClient(c echo.Context, id openapi_types.UUID) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	err := h.crmService.DeleteClient(c.Request().Context(), id, claims.UserID, claims.Role)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrClientNotFound):
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", "client not found")
		case errors.Is(err, crm.ErrUnauthorizedAction):
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete client")
		}
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ClientHandler) ArchiveClient(c echo.Context, id openapi_types.UUID) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can archive clients")
	}

	client, err := h.crmService.ArchiveClient(c.Request().Context(), id, claims.UserID, claims.Role)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrClientNotFound):
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", "client not found")
		case errors.Is(err, crm.ErrUnauthorizedAction):
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to archive client")
		}
	}

	return c.JSON(http.StatusOK, MapClientToResponse(client))
}

func (h *ClientHandler) UnarchiveClient(c echo.Context, id openapi_types.UUID) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can unarchive clients")
	}

	client, err := h.crmService.UnarchiveClient(c.Request().Context(), id, claims.UserID, claims.Role)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrClientNotFound):
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", "client not found")
		case errors.Is(err, crm.ErrUnauthorizedAction):
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to unarchive client")
		}
	}

	return c.JSON(http.StatusOK, MapClientToResponse(client))
}

