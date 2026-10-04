package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
)

// UserRepository определяет контракт доступа к данным пользователей.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}

// PasswordHasher определяет контракт для хэширования паролей.
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) error
}

// TokenManager определяет контракт для работы с JWT токенами.
type TokenManager interface {
	GenerateAccessToken(user *domain.User) (string, int, error)
	GenerateRefreshToken() (string, error)
	RefreshTTL() time.Duration
}

// SessionStore определяет контракт хранения refresh сессий.
type SessionStore interface {
	SaveRefreshToken(ctx context.Context, refreshToken string, userID uuid.UUID, ttl time.Duration) error
	GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (uuid.UUID, error)
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
}

type RegisterInput struct {
	Email    string
	Password string
	FullName string
	Phone    *string
	Role     domain.Role
}

type LoginInput struct {
	Email    string
	Password string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int
}

type AuthResult struct {
	User   *domain.User
	Tokens TokenPair
}

// Service реализует сценарии аутентификации и управления пользователями.
type Service struct {
	userRepo     UserRepository
	hasher       PasswordHasher
	tokenManager TokenManager
	sessionStore SessionStore
}

func NewService(
	userRepo UserRepository,
	hasher PasswordHasher,
	tokenManager TokenManager,
	sessionStore SessionStore,
) *Service {
	return &Service{
		userRepo:     userRepo,
		hasher:       hasher,
		tokenManager: tokenManager,
		sessionStore: sessionStore,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	cleanEmail, err := domain.ValidateEmail(input.Email)
	if err != nil {
		return nil, err
	}

	if err := domain.ValidatePassword(input.Password); err != nil {
		return nil, err
	}

	cleanName, err := domain.ValidateFullName(input.FullName)
	if err != nil {
		return nil, err
	}

	role := input.Role
	if role == "" {
		role = domain.RoleStudent
	}
	if !role.IsValid() {
		return nil, domain.ErrInvalidRole
	}

	// Проверяем существование пользователя по email
	existing, err := s.userRepo.GetByEmail(ctx, cleanEmail)
	if err == nil && existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	passwordHash, err := s.hasher.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	user := domain.User{
		ID:           uuid.New(),
		Email:        cleanEmail,
		PasswordHash: passwordHash,
		FullName:     cleanName,
		Phone:        input.Phone,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, &user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	tokens, err := s.generateTokens(ctx, &user)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:   &user,
		Tokens: *tokens,
	}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	cleanEmail, err := domain.ValidateEmail(input.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := s.userRepo.GetByEmail(ctx, cleanEmail)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := s.hasher.ComparePassword(user.PasswordHash, input.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	tokens, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:   user,
		Tokens: *tokens,
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	if refreshToken == "" {
		return nil, domain.ErrInvalidToken
	}

	userID, err := s.sessionStore.GetUserIDByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, domain.ErrSessionExpired
	}

	// Ротация токена: отзываем старый
	_ = s.sessionStore.RevokeRefreshToken(ctx, refreshToken)

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	tokens, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:   user,
		Tokens: *tokens,
	}, nil
}

func (s *Service) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (s *Service) generateTokens(ctx context.Context, user *domain.User) (*TokenPair, error) {
	accessToken, expiresIn, err := s.tokenManager.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := s.sessionStore.SaveRefreshToken(ctx, refreshToken, user.ID, s.tokenManager.RefreshTTL()); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}, nil
}
