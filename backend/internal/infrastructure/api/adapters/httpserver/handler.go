package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/C4erries/school/backend/internal/application/auth"
	"github.com/C4erries/school/backend/internal/application/crm"
	"github.com/C4erries/school/backend/internal/application/dashboard"
	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/security"
)

// APIHandler реализует сгенерированный generated.ServerInterface из OpenAPI спеки.
type APIHandler struct {
	authService      *auth.Service
	scheduleService  *schedule.Service
	crmService       *crm.Service
	dashboardService *dashboard.Service
	tokenValidator   TokenValidator
	version          string
}

func NewAPIHandler(
	authService *auth.Service,
	scheduleService *schedule.Service,
	crmService *crm.Service,
	dashboardService *dashboard.Service,
	tokenValidator TokenValidator,
	version string,
) *APIHandler {
	return &APIHandler{
		authService:      authService,
		scheduleService:  scheduleService,
		crmService:       crmService,
		dashboardService: dashboardService,
		tokenValidator:   tokenValidator,
		version:          version,
	}
}

func (h *APIHandler) authenticate(w http.ResponseWriter, r *http.Request) (*security.UserClaims, bool) {
	claims, ok := UserFromContext(r.Context())
	if ok && claims != nil {
		return claims, true
	}

	token := ExtractBearerToken(r)
	if token == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
		return nil, false
	}

	claims, err := h.tokenValidator.ValidateAccessToken(token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired access token")
		return nil, false
	}

	return claims, true
}

// GetHealth реализует эндпоинт GET /health из OpenAPI спецификации.
func (h *APIHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	resp := generated.HealthResponse{
		Status:    "ok",
		Service:   "school-api",
		Timestamp: time.Now().UTC(),
		Version:   &h.version,
	}
	writeJSON(w, http.StatusOK, resp)
}

// Register реализует эндпоинт POST /auth/register.
func (h *APIHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req generated.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	role := domain.RoleStudent
	if req.Role != nil {
		role = domain.Role(*req.Role)
	}

	result, err := h.authService.Register(r.Context(), auth.RegisterInput{
		Email:    string(req.Email),
		Password: req.Password,
		FullName: req.FullName,
		Phone:    req.Phone,
		Role:     role,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidEmail),
			errors.Is(err, domain.ErrInvalidPassword),
			errors.Is(err, domain.ErrInvalidFullName),
			errors.Is(err, domain.ErrInvalidRole):
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		case errors.Is(err, domain.ErrUserAlreadyExists):
			writeError(w, http.StatusConflict, "USER_EXISTS", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to register user")
		}
		return
	}

	resp := generated.AuthResponse{
		Tokens: generated.TokenPair{
			AccessToken:  result.Tokens.AccessToken,
			RefreshToken: result.Tokens.RefreshToken,
			TokenType:    result.Tokens.TokenType,
			ExpiresIn:    result.Tokens.ExpiresIn,
		},
		User: mapUserToResponse(result.User),
	}

	writeJSON(w, http.StatusCreated, resp)
}

// Login реализует эндпоинт POST /auth/login.
func (h *APIHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req generated.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	result, err := h.authService.Login(r.Context(), auth.LoginInput{
		Email:    string(req.Email),
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to login")
		return
	}

	resp := generated.AuthResponse{
		Tokens: generated.TokenPair{
			AccessToken:  result.Tokens.AccessToken,
			RefreshToken: result.Tokens.RefreshToken,
			TokenType:    result.Tokens.TokenType,
			ExpiresIn:    result.Tokens.ExpiresIn,
		},
		User: mapUserToResponse(result.User),
	}

	writeJSON(w, http.StatusOK, resp)
}

// RefreshTokens реализует эндпоинт POST /auth/refresh.
func (h *APIHandler) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	var req generated.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	result, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidToken), errors.Is(err, domain.ErrSessionExpired):
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired refresh token")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to refresh tokens")
		}
		return
	}

	resp := generated.AuthResponse{
		Tokens: generated.TokenPair{
			AccessToken:  result.Tokens.AccessToken,
			RefreshToken: result.Tokens.RefreshToken,
			TokenType:    result.Tokens.TokenType,
			ExpiresIn:    result.Tokens.ExpiresIn,
		},
		User: mapUserToResponse(result.User),
	}

	writeJSON(w, http.StatusOK, resp)
}

func mapUserToResponse(u *domain.User) generated.UserResponse {
	var rateIndiv, ratePair, rateGroup *float32
	if u.DefaultRateIndividual != nil {
		v := float32(*u.DefaultRateIndividual)
		rateIndiv = &v
	}
	if u.DefaultRatePair != nil {
		v := float32(*u.DefaultRatePair)
		ratePair = &v
	}
	if u.DefaultRateGroup != nil {
		v := float32(*u.DefaultRateGroup)
		rateGroup = &v
	}

	return generated.UserResponse{
		Id:                    u.ID,
		Email:                 openapi_types.Email(u.Email),
		FullName:              u.FullName,
		Phone:                 u.Phone,
		Role:                  generated.Role(u.Role),
		DefaultRateIndividual: rateIndiv,
		DefaultRatePair:       ratePair,
		DefaultRateGroup:      rateGroup,
		CreatedAt:             u.CreatedAt,
	}
}

// GetCurrentUser реализует эндпоинт GET /auth/me.
func (h *APIHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		// Если запрос пришел без middleware, проверяем заголовок вручную
		token := ExtractBearerToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
			return
		}
		var err error
		claims, err = h.tokenValidator.ValidateAccessToken(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
			return
		}
	}

	user, err := h.authService.GetCurrentUser(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get current user")
		return
	}

	writeJSON(w, http.StatusOK, mapUserToResponse(user))
}

// GetUserDefaultRates реализует GET /users/me/rates.
func (h *APIHandler) GetUserDefaultRates(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view default rates")
		return
	}

	rates, err := h.authService.GetDefaultRates(r.Context(), claims.UserID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get default rates")
		}
		return
	}

	writeJSON(w, http.StatusOK, generated.UserDefaultRates{
		RateIndividual: float32(rates.RateIndividual),
		RatePair:       float32(rates.RatePair),
		RateGroup:      float32(rates.RateGroup),
	})
}

// UpdateUserDefaultRates реализует PUT /users/me/rates.
func (h *APIHandler) UpdateUserDefaultRates(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can update default rates")
		return
	}

	var req generated.UserDefaultRates
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.RateIndividual < 0 || req.RatePair < 0 || req.RateGroup < 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "rates cannot be negative")
		return
	}

	err := h.authService.UpdateDefaultRates(r.Context(), claims.UserID, auth.UserDefaultRates{
		RateIndividual: float64(req.RateIndividual),
		RatePair:       float64(req.RatePair),
		RateGroup:      float64(req.RateGroup),
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update default rates")
		}
		return
	}

	writeJSON(w, http.StatusOK, req)
}

// Проверка на этапе компиляции, что APIHandler полностью имплементирует ServerInterface.
var _ generated.ServerInterface = (*APIHandler)(nil)
