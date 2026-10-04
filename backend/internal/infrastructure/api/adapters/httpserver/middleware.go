package httpserver

import (
	"context"
	"net/http"
	"strings"

	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/security"
)

type contextKey string

const userClaimsKey contextKey = "user_claims"

// TokenValidator валидирует JWT access-токен.
type TokenValidator interface {
	ValidateAccessToken(tokenString string) (*security.UserClaims, error)
}

// UserFromContext извлекает утверждения пользователя из контекста запроса.
func UserFromContext(ctx context.Context) (*security.UserClaims, bool) {
	claims, ok := ctx.Value(userClaimsKey).(*security.UserClaims)
	return claims, ok && claims != nil
}

// ContextWithUser помещает claims пользователя в контекст.
func ContextWithUser(ctx context.Context, claims *security.UserClaims) context.Context {
	return context.WithValue(ctx, userClaimsKey, claims)
}

// ExtractBearerToken извлекает Bearer токен из заголовка Authorization.
func ExtractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
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
func AuthMiddleware(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ExtractBearerToken(r)
			if token == "" {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing or invalid authorization header")
				return
			}

			claims, err := validator.ValidateAccessToken(token)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired access token")
				return
			}

			ctx := ContextWithUser(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRoles проверяет, что роль авторизованного пользователя входит в список разрешенных.
func RequireRoles(roles ...domain.Role) func(http.Handler) http.Handler {
	allowedRoles := make(map[domain.Role]struct{}, len(roles))
	for _, r := range roles {
		allowedRoles[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := UserFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
				return
			}

			if _, allowed := allowedRoles[claims.Role]; !allowed {
				writeError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
