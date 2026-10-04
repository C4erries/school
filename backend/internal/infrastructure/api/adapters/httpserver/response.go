package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver/generated"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	resp := generated.ErrorResponse{
		Error: generated.ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
	writeJSON(w, status, resp)
}
