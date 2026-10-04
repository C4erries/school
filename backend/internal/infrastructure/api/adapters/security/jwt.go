package security

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

// UserClaims содержит полезную нагрузку JWT access токена.
type UserClaims struct {
	UserID uuid.UUID   `json:"user_id"`
	Email  string      `json:"email"`
	Role   domain.Role `json:"role"`
	jwt.RegisteredClaims
}

// TokenManager инкапсулирует операции генерации и валидации JWT токенов.
type TokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewTokenManager(secret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (tm *TokenManager) AccessTTL() time.Duration {
	return tm.accessTTL
}

func (tm *TokenManager) RefreshTTL() time.Duration {
	return tm.refreshTTL
}

// GenerateAccessToken создает подписанный JWT токен для пользователя.
func (tm *TokenManager) GenerateAccessToken(user *domain.User) (string, int, error) {
	now := time.Now().UTC()
	claims := UserClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tm.accessTTL)),
			Issuer:    "school-platform",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(tm.secret)
	if err != nil {
		return "", 0, fmt.Errorf("sign access token: %w", err)
	}

	return tokenString, int(tm.accessTTL.Seconds()), nil
}

// ValidateAccessToken проверяет подпись и парсит утверждения из токена.
func (tm *TokenManager) ValidateAccessToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return tm.secret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GenerateRefreshToken генерирует случайный защищенный refresh-токен.
func (tm *TokenManager) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
