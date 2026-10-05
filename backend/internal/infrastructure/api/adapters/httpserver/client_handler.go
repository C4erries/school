package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver/generated"
)

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
		resp = append(resp, generated.ClientResponse{
			Id:               c.ID,
			TeacherId:        c.TeacherID,
			Name:             c.Name,
			Phone:            c.Phone,
			BaseRate:         float32(c.BaseRate),
			SchoolPercentTag: c.SchoolPercentTag,
			CreatedAt:        c.CreatedAt,
		})
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

	if req.BaseRate <= 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "base_rate must be greater than 0")
		return
	}

	schoolPercentTag := 0
	if req.SchoolPercentTag != nil {
		schoolPercentTag = *req.SchoolPercentTag
	}

	client, err := h.crmService.CreateClient(
		r.Context(),
		claims.UserID,
		name,
		req.Phone,
		float64(req.BaseRate),
		schoolPercentTag,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create client")
		return
	}

	writeJSON(w, http.StatusCreated, generated.ClientResponse{
		Id:               client.ID,
		TeacherId:        client.TeacherID,
		Name:             client.Name,
		Phone:            client.Phone,
		BaseRate:         float32(client.BaseRate),
		SchoolPercentTag: client.SchoolPercentTag,
		CreatedAt:        client.CreatedAt,
	})
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
			Type:      string(s.Type),
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

	subType := domain.SubscriptionType(req.Type)
	if subType != domain.SubscriptionTypeLessons && subType != domain.SubscriptionTypeHours {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid subscription type, must be 'lessons' or 'hours'")
		return
	}

	sub, err := h.crmService.CreateSubscription(r.Context(), id, subType, float64(req.Balance))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create subscription")
		return
	}

	writeJSON(w, http.StatusCreated, generated.SubscriptionResponse{
		Id:        sub.ID,
		ClientId:  sub.ClientID,
		Type:      string(sub.Type),
		Balance:   float32(sub.Balance),
		CreatedAt: sub.CreatedAt,
	})
}
