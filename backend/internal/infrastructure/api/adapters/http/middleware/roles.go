package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/response"
)

// RequireRoles проверяет, что роль авторизованного пользователя входит в список разрешенных.
func RequireRoles(roles ...domain.Role) echo.MiddlewareFunc {
	allowedRoles := make(map[domain.Role]struct{}, len(roles))
	for _, r := range roles {
		allowedRoles[r] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := UserFromContext(c)
			if !ok {
				return response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
			}

			if _, allowed := allowedRoles[claims.Role]; !allowed {
				return response.Error(c, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
			}

			return next(c)
		}
	}
}

