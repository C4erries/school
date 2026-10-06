package finance

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/C4erries/school/backend/internal/application/finance"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/auth"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/response"
)

// FinanceHandler обрабатывает операции по платежам, сводкам и выплатам партнерам.
type FinanceHandler struct {
	financeService *finance.Service
	authHandler    *auth.Handler
}

func NewFinanceHandler(financeService *finance.Service, authHandler *auth.Handler) *FinanceHandler {
	return &FinanceHandler{
		financeService: financeService,
		authHandler:    authHandler,
	}
}

func MapPaymentToResponse(it *finance.PaymentItem) generated.PaymentResponse {
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

func (h *FinanceHandler) GetFinanceSummary(c echo.Context, params generated.GetFinanceSummaryParams) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view finance summary")
	}

	if h.financeService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "finance service not configured")
	}

	var month string
	if params.Month != nil {
		month = *params.Month
	}

	summary, err := h.financeService.GetFinanceSummary(c.Request().Context(), claims.UserID, month)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPeriodMonth) {
			return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate finance summary")
	}

	return c.JSON(http.StatusOK, generated.FinanceSummaryResponse{
		Month:                    summary.Month,
		TotalPayments:            float32(summary.TotalPayments),
		TotalEarned:              float32(summary.TotalEarned),
		TotalDebts:               float32(summary.TotalDebts),
		TotalCommissions:         float32(summary.TotalCommissions),
		ActiveSubscriptionsCount: summary.ActiveSubscriptionsCount,
		DebtorsCount:             summary.DebtorsCount,
	})
}

func (h *FinanceHandler) ListPayments(c echo.Context, params generated.ListPaymentsParams) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can list payments")
	}

	if h.financeService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "finance service not configured")
	}

	filter := finance.PaymentFilter{
		TeacherID: &claims.UserID,
		ClientID:  params.ClientId,
		From:      params.From,
		To:        params.To,
		Limit:     params.Limit,
		Offset:    params.Offset,
	}

	items, err := h.financeService.ListPayments(c.Request().Context(), filter)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list payments")
	}

	resp := make([]generated.PaymentResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, MapPaymentToResponse(it))
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *FinanceHandler) CreatePayment(c echo.Context) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can create payments")
	}

	if h.financeService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "finance service not configured")
	}

	var req generated.CreatePaymentRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	method := domain.PaymentMethodTransfer
	if req.PaymentMethod != nil {
		method = domain.PaymentMethod(*req.PaymentMethod)
	}

	var notes string
	if req.Notes != nil {
		notes = *req.Notes
	}

	item, err := h.financeService.CreatePayment(c.Request().Context(), finance.CreatePaymentInput{
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
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", "client not found")
		}
		if errors.Is(err, finance.ErrUnauthorizedAction) {
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", err.Error())
		}
		if errors.Is(err, domain.ErrInvalidPaymentAmount) ||
			errors.Is(err, domain.ErrInvalidPaymentHours) ||
			errors.Is(err, domain.ErrInvalidPaymentFormat) ||
			errors.Is(err, domain.ErrInvalidPaymentMethod) {
			return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return c.JSON(http.StatusCreated, MapPaymentToResponse(item))
}

func (h *FinanceHandler) ListPartnerSettlements(c echo.Context, params generated.ListPartnerSettlementsParams) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view partner settlements")
	}

	if h.financeService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "finance service not configured")
	}

	var month string
	if params.Month != nil {
		month = *params.Month
	}

	items, err := h.financeService.ListPartnerSettlements(c.Request().Context(), claims.UserID, month)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPeriodMonth) {
			return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate partner settlements")
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

	return c.JSON(http.StatusOK, resp)
}

func (h *FinanceHandler) CreatePartnerPayout(c echo.Context) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can create partner payout")
	}

	if h.financeService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "finance service not configured")
	}

	var req generated.CreatePartnerPayoutRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	var notes string
	if req.Notes != nil {
		notes = *req.Notes
	}

	payout, err := h.financeService.CreatePartnerPayout(c.Request().Context(), finance.CreatePartnerPayoutInput{
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
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", "tag not found")
		}
		if errors.Is(err, finance.ErrUnauthorizedAction) {
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", err.Error())
		}
		if errors.Is(err, finance.ErrPayoutAlreadyExists) {
			return response.Error(c, http.StatusConflict, "CONFLICT", err.Error())
		}
		if errors.Is(err, domain.ErrInvalidPeriodMonth) ||
			errors.Is(err, domain.ErrInvalidGrossAmount) ||
			errors.Is(err, domain.ErrInvalidCommissionAmount) {
			return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return c.JSON(http.StatusCreated, generated.PartnerPayoutResponse{
		Id:               payout.ID,
		TeacherId:        payout.TeacherID,
		TagId:            payout.TagID,
		PeriodMonth:      payout.PeriodMonth,
		GrossAmount:      float32(payout.GrossAmount),
		CommissionAmount: float32(payout.CommissionAmount),
		PaidAt:           payout.PaidAt,
		Notes:            req.Notes,
		CreatedAt:        payout.CreatedAt,
	})
}

