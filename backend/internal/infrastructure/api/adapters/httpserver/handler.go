package httpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver/generated"
)

// APIHandler реализует сгенерированный generated.ServerInterface из OpenAPI спеки.
type APIHandler struct {
	version string
}

func NewAPIHandler(version string) *APIHandler {
	return &APIHandler{version: version}
}

// GetHealth реализует эндпоинт GET /health из OpenAPI спецификации.
func (h *APIHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	resp := generated.HealthResponse{
		Status:    "ok",
		Service:   "school-api",
		Timestamp: time.Now().UTC(),
		Version:   &h.version,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// Проверка на этапе компиляции, что APIHandler полностью имплементирует ServerInterface.
var _ generated.ServerInterface = (*APIHandler)(nil)
