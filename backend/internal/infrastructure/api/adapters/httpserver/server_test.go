package httpserver_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver"
)

func TestBuildMux_HealthEndpoints(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := httpserver.NewAPIHandler(nil, nil, nil, "v1")
	mux := httpserver.BuildMux(handler, logger)

	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
		checkBody      func(t *testing.T, body []byte)
	}{
		{
			name:           "GET /health returns 200 and ok status",
			path:           "/health",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]any
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				assert.Equal(t, "ok", resp["status"])
				assert.Equal(t, "school-api", resp["service"])
				assert.NotEmpty(t, resp["timestamp"])
			},
		},
		{
			name:           "GET /api/v1/health returns 200 and version",
			path:           "/api/v1/health",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]any
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				assert.Equal(t, "ok", resp["status"])
				assert.Equal(t, "school-api", resp["service"])
				assert.Equal(t, "v1", resp["version"])
			},
		},
		{
			name:           "POST /health returns 405 Method Not Allowed",
			path:           "/health",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			checkBody:      nil,
		},
		{
			name:           "GET /unknown-path returns 404 Not Found",
			path:           "/unknown-path",
			method:         http.MethodGet,
			expectedStatus: http.StatusNotFound,
			checkBody:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.checkBody != nil {
				tt.checkBody(t, rec.Body.Bytes())
			}
		})
	}
}
