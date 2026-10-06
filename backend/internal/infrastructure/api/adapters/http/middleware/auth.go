package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/response"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/security"
)

type contextKey string

const userClaimsContextKey contextKey = "user_claims"
const userClaimsEchoKey = "user_claims"

// TokenValidator валидирует JWT access-токен.
type TokenValidator interface {
	ValidateAccessToken(tokenString string) (*security.UserClaims, error)
}

// UserFromContext извлекает утверждения пользователя из контекста (echo.Context или context.Context).
func UserFromContext(ctx any) (*security.UserClaims, bool) {
	if c, ok := ctx.(echo.Context); ok {
		if claims, ok := c.Get(userClaimsEchoKey).(*security.UserClaims); ok && claims != nil {
			return claims, true
		}
		// Fallback to std context inside echo.Context
		if claims, ok := c.Request().Context().Value(userClaimsContextKey).(*security.UserClaims); ok && claims != nil {
			return claims, true
		}
		return nil, false
	}

	if stdCtx, ok := ctx.(context.Context); ok {
		if claims, ok := stdCtx.Value(userClaimsContextKey).(*security.UserClaims); ok && claims != nil {
			return claims, true
		}
	}

	return nil, false
}

// ContextWithUser помещает claims пользователя в context.Context.
func ContextWithUser(ctx context.Context, claims *security.UserClaims) context.Context {
	return context.WithValue(ctx, userClaimsContextKey, claims)
}

// ExtractBearerToken извлекает Bearer токен из заголовка Authorization.
func ExtractBearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// AuthMiddleware проверяет наличие и валидность Bearer токена.
func AuthMiddleware(validator TokenValidator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			token := ExtractBearerToken(authHeader)
			if token == "" {
				return response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing or invalid authorization header")
			}

			claims, err := validator.ValidateAccessToken(token)
			if err != nil {
				return response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired access token")
			}

			c.Set(userClaimsEchoKey, claims)
			ctx := ContextWithUser(c.Request().Context(), claims)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

