package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/application/auth"
	"github.com/C4erries/school/backend/internal/domain"
)

// MockUserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if u := args.Get(0); u != nil {
		return u.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if u := args.Get(0); u != nil {
		return u.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// MockPasswordHasher
type MockPasswordHasher struct {
	mock.Mock
}

func (m *MockPasswordHasher) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHasher) ComparePassword(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

// MockTokenManager
type MockTokenManager struct {
	mock.Mock
}

func (m *MockTokenManager) GenerateAccessToken(user *domain.User) (string, int, error) {
	args := m.Called(user)
	return args.String(0), args.Int(1), args.Error(2)
}

func (m *MockTokenManager) GenerateRefreshToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockTokenManager) RefreshTTL() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

// MockSessionStore
type MockSessionStore struct {
	mock.Mock
}

func (m *MockSessionStore) SaveRefreshToken(ctx context.Context, refreshToken string, userID uuid.UUID, ttl time.Duration) error {
	args := m.Called(ctx, refreshToken, userID, ttl)
	return args.Error(0)
}

func (m *MockSessionStore) GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (uuid.UUID, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockSessionStore) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		input       auth.RegisterInput
		setupMocks  func(r *MockUserRepository, h *MockPasswordHasher, tm *MockTokenManager, s *MockSessionStore)
		expectedErr error
	}{
		{
			name: "Success with default role",
			input: auth.RegisterInput{
				Email:    "test@school.ru",
				Password: "password123",
				FullName: "Иван Иванов",
			},
			setupMocks: func(r *MockUserRepository, h *MockPasswordHasher, tm *MockTokenManager, s *MockSessionStore) {
				r.On("GetByEmail", ctx, "test@school.ru").Return(nil, domain.ErrUserNotFound)
				h.On("HashPassword", "password123").Return("hashed_pw", nil)
				r.On("Create", ctx, mock.MatchedBy(func(u *domain.User) bool {
					return u.Email == "test@school.ru" && u.Role == domain.RoleStudent && u.FullName == "Иван Иванов"
				})).Return(nil)
				tm.On("GenerateAccessToken", mock.AnythingOfType("*domain.User")).Return("access_token", 900, nil)
				tm.On("GenerateRefreshToken").Return("refresh_token", nil)
				tm.On("RefreshTTL").Return(7 * 24 * time.Hour)
				s.On("SaveRefreshToken", ctx, "refresh_token", mock.AnythingOfType("uuid.UUID"), 7*24*time.Hour).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "Success with teacher role",
			input: auth.RegisterInput{
				Email:    "teacher@school.ru",
				Password: "password123",
				FullName: "Пётр Петров",
				Role:     domain.RoleTeacher,
			},
			setupMocks: func(r *MockUserRepository, h *MockPasswordHasher, tm *MockTokenManager, s *MockSessionStore) {
				r.On("GetByEmail", ctx, "teacher@school.ru").Return(nil, domain.ErrUserNotFound)
				h.On("HashPassword", "password123").Return("hashed_pw", nil)
				r.On("Create", ctx, mock.MatchedBy(func(u *domain.User) bool {
					return u.Role == domain.RoleTeacher
				})).Return(nil)
				tm.On("GenerateAccessToken", mock.AnythingOfType("*domain.User")).Return("access_token", 900, nil)
				tm.On("GenerateRefreshToken").Return("refresh_token", nil)
				tm.On("RefreshTTL").Return(7 * 24 * time.Hour)
				s.On("SaveRefreshToken", ctx, "refresh_token", mock.AnythingOfType("uuid.UUID"), 7*24*time.Hour).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "Fail on invalid email",
			input: auth.RegisterInput{
				Email:    "invalid-email",
				Password: "password123",
				FullName: "Иван Иванов",
			},
			setupMocks:  func(r *MockUserRepository, h *MockPasswordHasher, tm *MockTokenManager, s *MockSessionStore) {},
			expectedErr: domain.ErrInvalidEmail,
		},
		{
			name: "Fail on short password",
			input: auth.RegisterInput{
				Email:    "test@school.ru",
				Password: "123",
				FullName: "Иван Иванов",
			},
			setupMocks:  func(r *MockUserRepository, h *MockPasswordHasher, tm *MockTokenManager, s *MockSessionStore) {},
			expectedErr: domain.ErrInvalidPassword,
		},
		{
			name: "Fail on short name",
			input: auth.RegisterInput{
				Email:    "test@school.ru",
				Password: "password123",
				FullName: "A",
			},
			setupMocks:  func(r *MockUserRepository, h *MockPasswordHasher, tm *MockTokenManager, s *MockSessionStore) {},
			expectedErr: domain.ErrInvalidFullName,
		},
		{
			name: "Fail on invalid role",
			input: auth.RegisterInput{
				Email:    "test@school.ru",
				Password: "password123",
				FullName: "Иван Иванов",
				Role:     domain.Role("superadmin"),
			},
			setupMocks:  func(r *MockUserRepository, h *MockPasswordHasher, tm *MockTokenManager, s *MockSessionStore) {},
			expectedErr: domain.ErrInvalidRole,
		},
		{
			name: "Fail on user already exists",
			input: auth.RegisterInput{
				Email:    "test@school.ru",
				Password: "password123",
				FullName: "Иван Иванов",
			},
			setupMocks: func(r *MockUserRepository, h *MockPasswordHasher, tm *MockTokenManager, s *MockSessionStore) {
				r.On("GetByEmail", ctx, "test@school.ru").Return(&domain.User{ID: uuid.New(), Email: "test@school.ru"}, nil)
			},
			expectedErr: domain.ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(MockUserRepository)
			hasher := new(MockPasswordHasher)
			tokenMgr := new(MockTokenManager)
			sessionStore := new(MockSessionStore)

			tt.setupMocks(userRepo, hasher, tokenMgr, sessionStore)

			svc := auth.NewService(userRepo, hasher, tokenMgr, sessionStore)
			result, err := svc.Register(ctx, tt.input)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.NotEmpty(t, result.Tokens.AccessToken)
				assert.NotEmpty(t, result.Tokens.RefreshToken)
				assert.Equal(t, tt.input.Email, result.User.Email)
			}

			userRepo.AssertExpectations(t)
			hasher.AssertExpectations(t)
			tokenMgr.AssertExpectations(t)
			sessionStore.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()
	testUser := &domain.User{
		ID:           uuid.New(),
		Email:        "user@school.ru",
		PasswordHash: "hashed_pw",
		Role:         domain.RoleStudent,
		FullName:     "Студент",
	}

	tests := []struct {
		name        string
		input       auth.LoginInput
		setupMocks  func(r *MockUserRepository, h *MockPasswordHasher, tm *MockTokenManager, s *MockSessionStore)
		expectedErr error
	}{
		{
			name: "Success login",
			input: auth.LoginInput{
				Email:    "user@school.ru",
				Password: "correct_password",
			},
			setupMocks: func(r *MockUserRepository, h *MockPasswordHasher, tm *MockTokenManager, s *MockSessionStore) {
				r.On("GetByEmail", ctx, "user@school.ru").Return(testUser, nil)
				h.On("ComparePassword", "hashed_pw", "correct_password").Return(nil)
				tm.On("GenerateAccessToken", testUser).Return("access_token", 900, nil)
				tm.On("GenerateRefreshToken").Return("refresh_token", nil)
				tm.On("RefreshTTL").Return(7 * 24 * time.Hour)
				s.On("SaveRefreshToken", ctx, "refresh_token", testUser.ID, 7*24*time.Hour).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "Fail on wrong password",
			input: auth.LoginInput{
				Email:    "user@school.ru",
				Password: "wrong_password",
			},
			setupMocks: func(r *MockUserRepository, h *MockPasswordHasher, tm *MockTokenManager, s *MockSessionStore) {
				r.On("GetByEmail", ctx, "user@school.ru").Return(testUser, nil)
				h.On("ComparePassword", "hashed_pw", "wrong_password").Return(errors.New("mismatch"))
			},
			expectedErr: domain.ErrInvalidCredentials,
		},
		{
			name: "Fail on user not found",
			input: auth.LoginInput{
				Email:    "nonexistent@school.ru",
				Password: "password",
			},
			setupMocks: func(r *MockUserRepository, h *MockPasswordHasher, tm *MockTokenManager, s *MockSessionStore) {
				r.On("GetByEmail", ctx, "nonexistent@school.ru").Return(nil, domain.ErrUserNotFound)
			},
			expectedErr: domain.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(MockUserRepository)
			hasher := new(MockPasswordHasher)
			tokenMgr := new(MockTokenManager)
			sessionStore := new(MockSessionStore)

			tt.setupMocks(userRepo, hasher, tokenMgr, sessionStore)

			svc := auth.NewService(userRepo, hasher, tokenMgr, sessionStore)
			result, err := svc.Login(ctx, tt.input)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, testUser.Email, result.User.Email)
			}

			userRepo.AssertExpectations(t)
			hasher.AssertExpectations(t)
			tokenMgr.AssertExpectations(t)
			sessionStore.AssertExpectations(t)
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	ctx := context.Background()
	testUserID := uuid.New()
	testUser := &domain.User{
		ID:    testUserID,
		Email: "user@school.ru",
		Role:  domain.RoleTeacher,
	}

	tests := []struct {
		name        string
		token       string
		setupMocks  func(r *MockUserRepository, tm *MockTokenManager, s *MockSessionStore)
		expectedErr error
	}{
		{
			name:  "Success token rotation",
			token: "valid_old_refresh",
			setupMocks: func(r *MockUserRepository, tm *MockTokenManager, s *MockSessionStore) {
				s.On("GetUserIDByRefreshToken", ctx, "valid_old_refresh").Return(testUserID, nil)
				s.On("RevokeRefreshToken", ctx, "valid_old_refresh").Return(nil)
				r.On("GetByID", ctx, testUserID).Return(testUser, nil)
				tm.On("GenerateAccessToken", testUser).Return("new_access", 900, nil)
				tm.On("GenerateRefreshToken").Return("new_refresh", nil)
				tm.On("RefreshTTL").Return(7 * 24 * time.Hour)
				s.On("SaveRefreshToken", ctx, "new_refresh", testUserID, 7*24*time.Hour).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:  "Fail empty token",
			token: "",
			setupMocks: func(r *MockUserRepository, tm *MockTokenManager, s *MockSessionStore) {
			},
			expectedErr: domain.ErrInvalidToken,
		},
		{
			name:  "Fail expired session",
			token: "expired_token",
			setupMocks: func(r *MockUserRepository, tm *MockTokenManager, s *MockSessionStore) {
				s.On("GetUserIDByRefreshToken", ctx, "expired_token").Return(uuid.Nil, errors.New("not found"))
			},
			expectedErr: domain.ErrSessionExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(MockUserRepository)
			hasher := new(MockPasswordHasher)
			tokenMgr := new(MockTokenManager)
			sessionStore := new(MockSessionStore)

			tt.setupMocks(userRepo, tokenMgr, sessionStore)

			svc := auth.NewService(userRepo, hasher, tokenMgr, sessionStore)
			result, err := svc.RefreshToken(ctx, tt.token)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, "new_access", result.Tokens.AccessToken)
				assert.Equal(t, "new_refresh", result.Tokens.RefreshToken)
			}

			userRepo.AssertExpectations(t)
			tokenMgr.AssertExpectations(t)
			sessionStore.AssertExpectations(t)
		})
	}
}

func TestAuthService_GetCurrentUser(t *testing.T) {
	ctx := context.Background()
	testUserID := uuid.New()
	testUser := &domain.User{
		ID:    testUserID,
		Email: "user@school.ru",
		Role:  domain.RoleOwner,
	}

	userRepo := new(MockUserRepository)
	userRepo.On("GetByID", ctx, testUserID).Return(testUser, nil)

	svc := auth.NewService(userRepo, nil, nil, nil)
	user, err := svc.GetCurrentUser(ctx, testUserID)

	require.NoError(t, err)
	assert.Equal(t, testUser, user)
	userRepo.AssertExpectations(t)
}
