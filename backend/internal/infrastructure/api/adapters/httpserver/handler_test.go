package httpserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/application/auth"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/security"
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

func (m *MockTokenManager) ValidateAccessToken(tokenString string) (*security.UserClaims, error) {
	args := m.Called(tokenString)
	if claims := args.Get(0); claims != nil {
		return claims.(*security.UserClaims), args.Error(1)
	}
	return nil, args.Error(1)
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

func TestAPIHandler_Register(t *testing.T) {
	userRepo := new(MockUserRepository)
	hasher := new(MockPasswordHasher)
	tokenMgr := new(MockTokenManager)
	sessionStore := new(MockSessionStore)

	userRepo.On("GetByEmail", mock.Anything, "new@school.ru").Return(nil, domain.ErrUserNotFound)
	hasher.On("HashPassword", "password123").Return("hash", nil)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	tokenMgr.On("GenerateAccessToken", mock.AnythingOfType("*domain.User")).Return("access_token_123", 900, nil)
	tokenMgr.On("GenerateRefreshToken").Return("refresh_token_123", nil)
	tokenMgr.On("RefreshTTL").Return(7 * 24 * time.Hour)
	sessionStore.On("SaveRefreshToken", mock.Anything, "refresh_token_123", mock.Anything, 7*24*time.Hour).Return(nil)

	authSvc := auth.NewService(userRepo, hasher, tokenMgr, sessionStore)
	handler := httpserver.NewAPIHandler(authSvc, nil, tokenMgr, "v1")

	reqBody := `{"email":"new@school.ru","password":"password123","full_name":"Иван Иванов"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp generated.AuthResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "access_token_123", resp.Tokens.AccessToken)
	assert.Equal(t, "refresh_token_123", resp.Tokens.RefreshToken)
	assert.Equal(t, "new@school.ru", string(resp.User.Email))
}

func TestAPIHandler_Login_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	hasher := new(MockPasswordHasher)
	tokenMgr := new(MockTokenManager)
	sessionStore := new(MockSessionStore)

	uID := uuid.New()
	testUser := &domain.User{
		ID:           uID,
		Email:        "user@school.ru",
		PasswordHash: "valid_hash",
		Role:         domain.RoleStudent,
		FullName:     "Студент",
	}

	userRepo.On("GetByEmail", mock.Anything, "user@school.ru").Return(testUser, nil)
	hasher.On("ComparePassword", "valid_hash", "password123").Return(nil)
	tokenMgr.On("GenerateAccessToken", testUser).Return("access_token_login", 900, nil)
	tokenMgr.On("GenerateRefreshToken").Return("refresh_token_login", nil)
	tokenMgr.On("RefreshTTL").Return(7 * 24 * time.Hour)
	sessionStore.On("SaveRefreshToken", mock.Anything, "refresh_token_login", uID, 7*24*time.Hour).Return(nil)

	authSvc := auth.NewService(userRepo, hasher, tokenMgr, sessionStore)
	handler := httpserver.NewAPIHandler(authSvc, nil, tokenMgr, "v1")

	reqBody := `{"email":"user@school.ru","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp generated.AuthResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "access_token_login", resp.Tokens.AccessToken)
}

func TestAPIHandler_Login_InvalidCredentials(t *testing.T) {
	userRepo := new(MockUserRepository)
	hasher := new(MockPasswordHasher)
	tokenMgr := new(MockTokenManager)
	sessionStore := new(MockSessionStore)

	userRepo.On("GetByEmail", mock.Anything, "user@school.ru").Return(nil, domain.ErrUserNotFound)

	authSvc := auth.NewService(userRepo, hasher, tokenMgr, sessionStore)
	handler := httpserver.NewAPIHandler(authSvc, nil, tokenMgr, "v1")

	reqBody := `{"email":"user@school.ru","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAPIHandler_GetCurrentUser(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenMgr := new(MockTokenManager)

	uID := uuid.New()
	testUser := &domain.User{
		ID:        uID,
		Email:     "me@school.ru",
		Role:      domain.RoleTeacher,
		FullName:  "Учитель",
		CreatedAt: time.Now().UTC(),
	}

	claims := &security.UserClaims{
		UserID: uID,
		Email:  "me@school.ru",
		Role:   domain.RoleTeacher,
	}

	userRepo.On("GetByID", mock.Anything, uID).Return(testUser, nil)
	tokenMgr.On("ValidateAccessToken", "valid_jwt_token").Return(claims, nil)

	authSvc := auth.NewService(userRepo, nil, nil, nil)
	handler := httpserver.NewAPIHandler(authSvc, nil, tokenMgr, "v1")

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer valid_jwt_token")
	rec := httptest.NewRecorder()

	handler.GetCurrentUser(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp generated.UserResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, uID, resp.Id)
	assert.Equal(t, "me@school.ru", string(resp.Email))
	assert.Equal(t, generated.Role("teacher"), resp.Role)
}

func TestAuthMiddleware(t *testing.T) {
	tokenMgr := new(MockTokenManager)
	claims := &security.UserClaims{
		UserID: uuid.New(),
		Email:  "user@school.ru",
		Role:   domain.RoleStudent,
	}

	tokenMgr.On("ValidateAccessToken", "valid_token").Return(claims, nil)
	tokenMgr.On("ValidateAccessToken", "invalid_token").Return(nil, errors.New("invalid"))

	middleware := httpserver.AuthMiddleware(tokenMgr)

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		c, ok := httpserver.UserFromContext(r.Context())
		assert.True(t, ok)
		assert.Equal(t, claims.UserID, c.UserID)
		w.WriteHeader(http.StatusOK)
	})

	// Test 1: Valid token
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	rec := httptest.NewRecorder()
	middleware(nextHandler).ServeHTTP(rec, req)
	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Test 2: Missing header
	nextCalled = false
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	rec = httptest.NewRecorder()
	middleware(nextHandler).ServeHTTP(rec, req)
	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Test 3: Invalid token
	nextCalled = false
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	rec = httptest.NewRecorder()
	middleware(nextHandler).ServeHTTP(rec, req)
	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
