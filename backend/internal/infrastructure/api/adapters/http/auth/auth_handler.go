package auth

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/C4erries/school/backend/internal/application/auth"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/middleware"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/response"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/security"
)

// Handler обрабатывает HTTP запросы аутентификации и профиля пользователя.
type Handler struct {
	authService    *auth.Service
	tokenValidator middleware.TokenValidator
}

func NewHandler(authService *auth.Service, tokenValidator middleware.TokenValidator) *Handler {
	return &Handler{
		authService:    authService,
		tokenValidator: tokenValidator,
	}
}

func (h *Handler) Authenticate(c echo.Context) (*security.UserClaims, bool) {
	claims, ok := middleware.UserFromContext(c)
	if ok && claims != nil {
		return claims, true
	}

	authHeader := c.Request().Header.Get("Authorization")
	token := middleware.ExtractBearerToken(authHeader)
	if token == "" {
		_ = response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
		return nil, false
	}

	claims, err := h.tokenValidator.ValidateAccessToken(token)
	if err != nil {
		_ = response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired access token")
		return nil, false
	}

	return claims, true
}

func (h *Handler) Register(c echo.Context) error {
	var req generated.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	role := domain.RoleStudent
	if req.Role != nil {
		role = domain.Role(*req.Role)
	}

	result, err := h.authService.Register(c.Request().Context(), auth.RegisterInput{
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
			return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		case errors.Is(err, domain.ErrUserAlreadyExists):
			return response.Error(c, http.StatusConflict, "USER_EXISTS", err.Error())
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to register user")
		}
	}

	resp := generated.AuthResponse{
		Tokens: generated.TokenPair{
			AccessToken:  result.Tokens.AccessToken,
			RefreshToken: result.Tokens.RefreshToken,
			TokenType:    result.Tokens.TokenType,
			ExpiresIn:    result.Tokens.ExpiresIn,
		},
		User: MapUserToResponse(result.User),
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *Handler) Login(c echo.Context) error {
	var req generated.LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	result, err := h.authService.Login(c.Request().Context(), auth.LoginInput{
		Email:    string(req.Email),
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return response.Error(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to login")
	}

	resp := generated.AuthResponse{
		Tokens: generated.TokenPair{
			AccessToken:  result.Tokens.AccessToken,
			RefreshToken: result.Tokens.RefreshToken,
			TokenType:    result.Tokens.TokenType,
			ExpiresIn:    result.Tokens.ExpiresIn,
		},
		User: MapUserToResponse(result.User),
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) RefreshTokens(c echo.Context) error {
	var req generated.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	result, err := h.authService.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidToken), errors.Is(err, domain.ErrSessionExpired):
			return response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired refresh token")
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to refresh tokens")
		}
	}

	resp := generated.AuthResponse{
		Tokens: generated.TokenPair{
			AccessToken:  result.Tokens.AccessToken,
			RefreshToken: result.Tokens.RefreshToken,
			TokenType:    result.Tokens.TokenType,
			ExpiresIn:    result.Tokens.ExpiresIn,
		},
		User: MapUserToResponse(result.User),
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetCurrentUser(c echo.Context) error {
	claims, ok := h.Authenticate(c)
	if !ok {
		return nil
	}

	user, err := h.authService.GetCurrentUser(c.Request().Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", "user not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get current user")
	}

	return c.JSON(http.StatusOK, MapUserToResponse(user))
}

func (h *Handler) GetUserDefaultRates(c echo.Context) error {
	claims, ok := h.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can view default rates")
	}

	rates, err := h.authService.GetDefaultRates(c.Request().Context(), claims.UserID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", "user not found")
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get default rates")
		}
	}

	return c.JSON(http.StatusOK, generated.UserDefaultRates{
		RateIndividual: float32(rates.RateIndividual),
		RatePair:       float32(rates.RatePair),
		RateGroup:      float32(rates.RateGroup),
	})
}

func (h *Handler) UpdateUserDefaultRates(c echo.Context) error {
	claims, ok := h.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleTeacher && claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only teacher or admin can update default rates")
	}

	var req generated.UserDefaultRates
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	if req.RateIndividual < 0 || req.RatePair < 0 || req.RateGroup < 0 {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "rates cannot be negative")
	}

	err := h.authService.UpdateDefaultRates(c.Request().Context(), claims.UserID, auth.UserDefaultRates{
		RateIndividual: float64(req.RateIndividual),
		RatePair:       float64(req.RatePair),
		RateGroup:      float64(req.RateGroup),
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", "user not found")
		default:
			return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update default rates")
		}
	}

	return c.JSON(http.StatusOK, req)
}

func MapUserToResponse(u *domain.User) generated.UserResponse {
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

