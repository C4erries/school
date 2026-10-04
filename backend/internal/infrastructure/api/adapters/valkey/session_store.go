package valkey

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valkey-io/valkey-go"
)

var ErrSessionNotFound = errors.New("refresh session not found or expired")

// SessionStore сохраняет и валидирует refresh-токены в Valkey.
type SessionStore struct {
	client valkey.Client
}

func NewSessionStore(client valkey.Client) *SessionStore {
	return &SessionStore{client: client}
}

func refreshKey(token string) string {
	return fmt.Sprintf("session:refresh:%s", token)
}

// SaveRefreshToken сохраняет refresh token с привязкой к ID пользователя и TTL.
func (s *SessionStore) SaveRefreshToken(ctx context.Context, refreshToken string, userID uuid.UUID, ttl time.Duration) error {
	cmd := s.client.B().Set().
		Key(refreshKey(refreshToken)).
		Value(userID.String()).
		Ex(ttl).
		Build()

	if err := s.client.Do(ctx, cmd).Error(); err != nil {
		return fmt.Errorf("valkey set refresh token: %w", err)
	}
	return nil
}

// GetUserIDByRefreshToken возвращает ID пользователя по refresh токену.
func (s *SessionStore) GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (uuid.UUID, error) {
	cmd := s.client.B().Get().Key(refreshKey(refreshToken)).Build()
	val, err := s.client.Do(ctx, cmd).ToString()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return uuid.Nil, ErrSessionNotFound
		}
		return uuid.Nil, fmt.Errorf("valkey get refresh token: %w", err)
	}

	uid, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse user id from session: %w", err)
	}

	return uid, nil
}

// RevokeRefreshToken удаляет refresh токен при ротации или выходе из системы.
func (s *SessionStore) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	cmd := s.client.B().Del().Key(refreshKey(refreshToken)).Build()
	if err := s.client.Do(ctx, cmd).Error(); err != nil {
		return fmt.Errorf("valkey revoke refresh token: %w", err)
	}
	return nil
}
