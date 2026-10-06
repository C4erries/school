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

// SubscriptionHandler обрабатывает подписки и ручные корректировки баланса.
type SubscriptionHandler struct {
	crmService  *crm.Service
	authHandler *auth.Handler
}

func NewSubscriptionHandler(crmService *crm.Service, authHandler *auth.Handler) *SubscriptionHandler {
	return &SubscriptionHandler{
		crmService:  crmService,
		authHandler: authHandler,
	}
}

func (h *SubscriptionHandler) ListSubscriptions(c echo.Context, id openapi_types.UUID) error {
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

	subs, err := h.crmService.ListSubscriptions(c.Request().Context(), id)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list subscriptions")
	}

	resp := make([]generated.SubscriptionResponse, 0, len(subs))
	for _, s := range subs {
		resp = append(resp, generated.SubscriptionResponse{
			Id:        s.ID,
			ClientId:  s.ClientID,
			Format:    generated.SubscriptionResponseFormat(s.Format),
			Balance:   float32(s.Balance),
			CreatedAt: s.CreatedAt,
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *SubscriptionHandler) CreateSubscription(c echo.Context, id openapi_types.UUID) error {
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

	var req generated.CreateSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	var formatStr string
	if req.Format != nil {
		formatStr = string(*req.Format)
	} else if req.Type != nil {
		formatStr = *req.Type
	}

	subFormat := domain.SubscriptionFormat(formatStr)
	if !subFormat.IsValid() {
		if formatStr == "lessons" || formatStr == "hours" {
			subFormat = domain.SubscriptionFormatIndividual
		} else {
			return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid subscription format, must be individual, pair, or group")
		}
	}

	sub, err := h.crmService.CreateSubscription(c.Request().Context(), id, subFormat, float64(req.Balance))
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create subscription")
	}

	return c.JSON(http.StatusCreated, generated.SubscriptionResponse{
		Id:        sub.ID,
		ClientId:  sub.ClientID,
		Format:    generated.SubscriptionResponseFormat(sub.Format),
		Balance:   float32(sub.Balance),
		CreatedAt: sub.CreatedAt,
	})
}

func (h *SubscriptionHandler) AdjustClientBalance(c echo.Context, id openapi_types.UUID) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can adjust client balance")
	}

	var req generated.AdjustBalanceRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "reason is required")
	}

	if req.DeltaHours == 0 {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "delta_hours cannot be zero")
	}

	format := domain.SubscriptionFormat(req.Format)
	if !format.IsValid() {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid subscription format")
	}

	client, err := h.crmService.AdjustBalance(c.Request().Context(), crm.AdjustBalanceInput{
		ClientID:   id,
		CallerID:   claims.UserID,
		CallerRole: claims.Role,
		Format:     format,
		DeltaHours: float64(req.DeltaHours),
		Reason:     reason,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrClientNotFound):
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", "client not found")
		case errors.Is(err, crm.ErrUnauthorizedAction):
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to adjust client balance")
		}
	}

	return c.JSON(http.StatusOK, MapClientToResponse(client))
}

