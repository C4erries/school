package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/C4erries/school/backend/internal/application/crm"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver/generated"
)

func mapClientToResponse(c *domain.Client) generated.ClientResponse {
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
		Tags:             tags,
		Balances:         balances,
		CreatedAt:        c.CreatedAt,
	}
}

// ListClients реализует GET /clients.
func (h *APIHandler) ListClients(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can list clients")
		return
	}

	clients, err := h.crmService.ListClients(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list clients")
		return
	}

	resp := make([]generated.ClientResponse, 0, len(clients))
	for _, c := range clients {
		resp = append(resp, mapClientToResponse(c))
	}

	writeJSON(w, http.StatusOK, resp)
}

// CreateClient реализует POST /clients.
func (h *APIHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can create clients")
		return
	}

	var req generated.CreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "client name is required")
		return
	}

	var rateIndividual float64
	if req.RateIndividual != nil && *req.RateIndividual > 0 {
		rateIndividual = float64(*req.RateIndividual)
	} else if req.BaseRate != nil && *req.BaseRate > 0 {
		rateIndividual = float64(*req.BaseRate)
	} else {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "rate_individual must be greater than 0")
		return
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

	client, err := h.crmService.CreateClientWithRates(r.Context(), crm.CreateClientInput{
		TeacherID:        claims.UserID,
		Name:             name,
		Phone:            req.Phone,
		BaseRate:         rateIndividual,
		RateIndividual:   rateIndividual,
		RatePair:         ratePair,
		RateGroup:        rateGroup,
		SchoolPercentTag: schoolPercentTag,
		TagIDs:           tagIDs,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create client")
		return
	}

	writeJSON(w, http.StatusCreated, mapClientToResponse(client))
}

// GetClient реализует GET /clients/{id}.
func (h *APIHandler) GetClient(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	client, err := h.crmService.GetClient(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrClientNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "client not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get client")
		return
	}

	if claims.Role != domain.RoleOwner && client.TeacherID != claims.UserID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		return
	}

	writeJSON(w, http.StatusOK, mapClientToResponse(client))
}

// UpdateClient реализует PATCH /clients/{id}.
func (h *APIHandler) UpdateClient(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	var req generated.UpdateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	var rateIndiv, ratePair, rateGroup *float64
	if req.RateIndividual != nil {
		v := float64(*req.RateIndividual)
		rateIndiv = &v
	}
	if req.RatePair != nil {
		v := float64(*req.RatePair)
		ratePair = &v
	}
	if req.RateGroup != nil {
		v := float64(*req.RateGroup)
		rateGroup = &v
	}

	client, err := h.crmService.UpdateClient(r.Context(), crm.UpdateClientInput{
		ID:             id,
		CallerID:       claims.UserID,
		CallerRole:     claims.Role,
		Name:           req.Name,
		Phone:          req.Phone,
		RateIndividual: rateIndiv,
		RatePair:       ratePair,
		RateGroup:      rateGroup,
		TagIDs:         req.TagIds,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrClientNotFound):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "client not found")
		case errors.Is(err, crm.ErrUnauthorizedAction):
			writeError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update client")
		}
		return
	}

	writeJSON(w, http.StatusOK, mapClientToResponse(client))
}

// DeleteClient реализует DELETE /clients/{id}.
func (h *APIHandler) DeleteClient(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	err := h.crmService.DeleteClient(r.Context(), id, claims.UserID, claims.Role)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrClientNotFound):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "client not found")
		case errors.Is(err, crm.ErrUnauthorizedAction):
			writeError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete client")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AssignClientTag реализует POST /clients/{id}/tags.
func (h *APIHandler) AssignClientTag(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	var req generated.AssignTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	client, err := h.crmService.AssignTagToClient(r.Context(), id, req.TagId, claims.UserID, claims.Role)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrClientNotFound), errors.Is(err, domain.ErrTagNotFound):
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		case errors.Is(err, crm.ErrUnauthorizedAction), errors.Is(err, domain.ErrUnauthorizedTagAction):
			writeError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to assign tag")
		}
		return
	}

	writeJSON(w, http.StatusOK, mapClientToResponse(client))
}

// RemoveClientTag реализует DELETE /clients/{id}/tags/{tag_id}.
func (h *APIHandler) RemoveClientTag(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, tagId openapi_types.UUID) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	client, err := h.crmService.RemoveTagFromClient(r.Context(), id, tagId, claims.UserID, claims.Role)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrClientNotFound):
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		case errors.Is(err, crm.ErrUnauthorizedAction):
			writeError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to remove tag")
		}
		return
	}

	writeJSON(w, http.StatusOK, mapClientToResponse(client))
}

// ListSubscriptions реализует GET /clients/{id}/subscriptions.
func (h *APIHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	client, err := h.crmService.GetClient(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrClientNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "client not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get client")
		return
	}

	if claims.Role != domain.RoleOwner && client.TeacherID != claims.UserID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		return
	}

	subs, err := h.crmService.ListSubscriptions(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list subscriptions")
		return
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

	writeJSON(w, http.StatusOK, resp)
}

// CreateSubscription реализует POST /clients/{id}/subscriptions.
func (h *APIHandler) CreateSubscription(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	client, err := h.crmService.GetClient(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrClientNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "client not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get client")
		return
	}

	if claims.Role != domain.RoleOwner && client.TeacherID != claims.UserID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		return
	}

	var req generated.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	var formatStr string
	if req.Format != nil {
		formatStr = string(*req.Format)
	} else if req.Type != nil {
		formatStr = *req.Type
	}

	subFormat := domain.SubscriptionFormat(formatStr)
	if !subFormat.IsValid() {
		// Для обратной совместимости, если кто-то передал 'lessons' или 'hours', преобразуем в 'individual'
		if formatStr == "lessons" || formatStr == "hours" {
			subFormat = domain.SubscriptionFormatIndividual
		} else {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid subscription format, must be individual, pair, or group")
			return
		}
	}

	sub, err := h.crmService.CreateSubscription(r.Context(), id, subFormat, float64(req.Balance))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create subscription")
		return
	}

	writeJSON(w, http.StatusCreated, generated.SubscriptionResponse{
		Id:        sub.ID,
		ClientId:  sub.ClientID,
		Format:    generated.SubscriptionResponseFormat(sub.Format),
		Balance:   float32(sub.Balance),
		CreatedAt: sub.CreatedAt,
	})
}
