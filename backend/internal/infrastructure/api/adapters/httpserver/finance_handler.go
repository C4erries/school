package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/C4erries/school/backend/internal/application/finance"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver/generated"
)

func mapPaymentToResponse(it *finance.PaymentItem) generated.PaymentResponse {
	p := it.Payment
	var clientName *string
	if it.ClientName != "" {
		clientName = &it.ClientName
	}
	var notes *string
	if p.Notes != "" {
		notes = &p.Notes
	}

	return generated.PaymentResponse{
		Id:            p.ID,
		TeacherId:     p.TeacherID,
		ClientId:      p.ClientID,
		ClientName:    clientName,
		Amount:        float32(p.Amount),
		Hours:         float32(p.Hours),
		Format:        generated.PaymentResponseFormat(p.Format),
		PaymentMethod: generated.PaymentResponsePaymentMethod(p.PaymentMethod),
		PaidAt:        p.PaidAt,
		Notes:         notes,
		CreatedAt:     p.CreatedAt,
	}
}

// GetFinanceSummary реализует GET /finance/summary.
func (h *APIHandler) GetFinanceSummary(w http.ResponseWriter, r *http.Request, params generated.GetFinanceSummaryParams) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view finance summary")
		return
	}

	if h.financeService == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "finance service not configured")
		return
	}

	var month string
	if params.Month != nil {
		month = *params.Month
	}

	summary, err := h.financeService.GetFinanceSummary(r.Context(), claims.UserID, month)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPeriodMonth) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate finance summary")
		return
	}

	resp := generated.FinanceSummaryResponse{
		Month:                    summary.Month,
		TotalPayments:            float32(summary.TotalPayments),
		TotalEarned:              float32(summary.TotalEarned),
		TotalDebts:               float32(summary.TotalDebts),
		TotalCommissions:         float32(summary.TotalCommissions),
		ActiveSubscriptionsCount: summary.ActiveSubscriptionsCount,
		DebtorsCount:             summary.DebtorsCount,
	}

	writeJSON(w, http.StatusOK, resp)
}

// ListPayments реализует GET /finance/payments.
func (h *APIHandler) ListPayments(w http.ResponseWriter, r *http.Request, params generated.ListPaymentsParams) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can list payments")
		return
	}

	if h.financeService == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "finance service not configured")
		return
	}

	filter := finance.PaymentFilter{
		TeacherID: &claims.UserID,
		ClientID:  params.ClientId,
		From:      params.From,
		To:        params.To,
		Limit:     params.Limit,
		Offset:    params.Offset,
	}

	items, err := h.financeService.ListPayments(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list payments")
		return
	}

	resp := make([]generated.PaymentResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, mapPaymentToResponse(it))
	}

	writeJSON(w, http.StatusOK, resp)
}

// CreatePayment реализует POST /finance/payments.
func (h *APIHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can create payments")
		return
	}

	if h.financeService == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "finance service not configured")
		return
	}

	var req generated.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	method := domain.PaymentMethodTransfer
	if req.PaymentMethod != nil {
		method = domain.PaymentMethod(*req.PaymentMethod)
	}

	var notes string
	if req.Notes != nil {
		notes = *req.Notes
	}

	item, err := h.financeService.CreatePayment(r.Context(), finance.CreatePaymentInput{
		TeacherID:     claims.UserID,
		ClientID:      req.ClientId,
		Amount:        float64(req.Amount),
		Hours:         float64(req.Hours),
		Format:        domain.SubscriptionFormat(req.Format),
		PaymentMethod: method,
		PaidAt:        req.PaidAt,
		Notes:         notes,
		CallerRole:    claims.Role,
	})
	if err != nil {
		if errors.Is(err, domain.ErrClientNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "client not found")
			return
		}
		if errors.Is(err, finance.ErrUnauthorizedAction) {
			writeError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		if errors.Is(err, domain.ErrInvalidPaymentAmount) ||
			errors.Is(err, domain.ErrInvalidPaymentHours) ||
			errors.Is(err, domain.ErrInvalidPaymentFormat) ||
			errors.Is(err, domain.ErrInvalidPaymentMethod) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, mapPaymentToResponse(item))
}

// ListPartnerSettlements реализует GET /finance/partner-settlements.
func (h *APIHandler) ListPartnerSettlements(w http.ResponseWriter, r *http.Request, params generated.ListPartnerSettlementsParams) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view partner settlements")
		return
	}

	if h.financeService == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "finance service not configured")
		return
	}

	var month string
	if params.Month != nil {
		month = *params.Month
	}

	items, err := h.financeService.ListPartnerSettlements(r.Context(), claims.UserID, month)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPeriodMonth) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate partner settlements")
		return
	}

	resp := make([]generated.PartnerSettlementResponse, 0, len(items))
	for _, it := range items {
		var color *string
		if it.TagColor != "" {
			color = &it.TagColor
		}
		resp = append(resp, generated.PartnerSettlementResponse{
			TagId:            it.TagID,
			TagName:          it.TagName,
			TagColor:         color,
			SchoolPercent:    it.SchoolPercent,
			PeriodMonth:      it.PeriodMonth,
			LessonsCount:     it.LessonsCount,
			GrossAmount:      float32(it.GrossAmount),
			CommissionAmount: float32(it.CommissionAmount),
			IsPaid:           it.IsPaid,
			PaidAt:           it.PaidAt,
			PayoutId:         it.PayoutID,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// CreatePartnerPayout реализует POST /finance/partner-payouts.
func (h *APIHandler) CreatePartnerPayout(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can create partner payout")
		return
	}

	if h.financeService == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "finance service not configured")
		return
	}

	var req generated.CreatePartnerPayoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	var notes string
	if req.Notes != nil {
		notes = *req.Notes
	}

	payout, err := h.financeService.CreatePartnerPayout(r.Context(), finance.CreatePartnerPayoutInput{
		TeacherID:        claims.UserID,
		TagID:            req.TagId,
		PeriodMonth:      req.PeriodMonth,
		GrossAmount:      float64(req.GrossAmount),
		CommissionAmount: float64(req.CommissionAmount),
		PaidAt:           req.PaidAt,
		Notes:            notes,
		CallerRole:       claims.Role,
	})
	if err != nil {
		if errors.Is(err, domain.ErrTagNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "tag not found")
			return
		}
		if errors.Is(err, finance.ErrUnauthorizedAction) {
			writeError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		if errors.Is(err, finance.ErrPayoutAlreadyExists) {
			writeError(w, http.StatusConflict, "CONFLICT", err.Error())
			return
		}
		if errors.Is(err, domain.ErrInvalidPeriodMonth) ||
			errors.Is(err, domain.ErrInvalidGrossAmount) ||
			errors.Is(err, domain.ErrInvalidCommissionAmount) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	resp := generated.PartnerPayoutResponse{
		Id:               payout.ID,
		TeacherId:        payout.TeacherID,
		TagId:            payout.TagID,
		PeriodMonth:      payout.PeriodMonth,
		GrossAmount:      float32(payout.GrossAmount),
		CommissionAmount: float32(payout.CommissionAmount),
		PaidAt:           payout.PaidAt,
		Notes:            req.Notes,
		CreatedAt:        payout.CreatedAt,
	}

	writeJSON(w, http.StatusCreated, resp)
}
