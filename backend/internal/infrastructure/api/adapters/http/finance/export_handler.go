package finance

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/C4erries/school/backend/internal/application/finance"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/auth"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/response"
)

// ExportHandler обрабатывает выгрузку отчетов в формате CSV (UTF-8 BOM).
type ExportHandler struct {
	financeService *finance.Service
	authHandler    *auth.Handler
}

func NewExportHandler(financeService *finance.Service, authHandler *auth.Handler) *ExportHandler {
	return &ExportHandler{
		financeService: financeService,
		authHandler:    authHandler,
	}
}

func (h *ExportHandler) ExportClients(c echo.Context) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can export clients")
	}

	if h.financeService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "finance service not configured")
	}

	data, err := h.financeService.ExportClientsCSV(c.Request().Context(), claims.UserID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to export clients")
	}

	filename := fmt.Sprintf("clients_%s.csv", time.Now().Format("2006-01-02"))
	c.Response().Header().Set("Content-Type", "text/csv; charset=utf-8")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	return c.Blob(http.StatusOK, "text/csv; charset=utf-8", data)
}

func (h *ExportHandler) ExportLessons(c echo.Context, params generated.ExportLessonsParams) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can export lessons")
	}

	if h.financeService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "finance service not configured")
	}

	data, err := h.financeService.ExportLessonsCSV(c.Request().Context(), claims.UserID, params.From, params.To)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to export lessons")
	}

	filename := fmt.Sprintf("lessons_%s.csv", time.Now().Format("2006-01-02"))
	c.Response().Header().Set("Content-Type", "text/csv; charset=utf-8")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	return c.Blob(http.StatusOK, "text/csv; charset=utf-8", data)
}

func (h *ExportHandler) ExportPayments(c echo.Context, params generated.ExportPaymentsParams) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can export payments")
	}

	if h.financeService == nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "finance service not configured")
	}

	data, err := h.financeService.ExportPaymentsCSV(c.Request().Context(), claims.UserID, params.From, params.To)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to export payments")
	}

	filename := fmt.Sprintf("payments_%s.csv", time.Now().Format("2006-01-02"))
	c.Response().Header().Set("Content-Type", "text/csv; charset=utf-8")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	return c.Blob(http.StatusOK, "text/csv; charset=utf-8", data)
}

