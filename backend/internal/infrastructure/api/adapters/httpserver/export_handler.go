package httpserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver/generated"
)

// ExportClients реализует GET /export/clients.
func (h *APIHandler) ExportClients(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can export clients")
		return
	}

	if h.financeService == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "finance service not configured")
		return
	}

	data, err := h.financeService.ExportClientsCSV(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to export clients")
		return
	}

	filename := fmt.Sprintf("clients_%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// ExportLessons реализует GET /export/lessons.
func (h *APIHandler) ExportLessons(w http.ResponseWriter, r *http.Request, params generated.ExportLessonsParams) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can export lessons")
		return
	}

	if h.financeService == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "finance service not configured")
		return
	}

	data, err := h.financeService.ExportLessonsCSV(r.Context(), claims.UserID, params.From, params.To)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to export lessons")
		return
	}

	filename := fmt.Sprintf("lessons_%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// ExportPayments реализует GET /export/payments.
func (h *APIHandler) ExportPayments(w http.ResponseWriter, r *http.Request, params generated.ExportPaymentsParams) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can export payments")
		return
	}

	if h.financeService == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "finance service not configured")
		return
	}

	data, err := h.financeService.ExportPaymentsCSV(r.Context(), claims.UserID, params.From, params.To)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to export payments")
		return
	}

	filename := fmt.Sprintf("payments_%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
