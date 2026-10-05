package httpserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/application/crm"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/security"
)

type MockCRMClientRepo struct {
	mock.Mock
}

func (m *MockCRMClientRepo) Create(ctx context.Context, client *domain.Client) error {
	return m.Called(ctx, client).Error(0)
}
func (m *MockCRMClientRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	args := m.Called(ctx, id)
	if c := args.Get(0); c != nil {
		return c.(*domain.Client), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockCRMClientRepo) ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.Client, error) {
	args := m.Called(ctx, teacherID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Client), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockCRMClientRepo) Update(ctx context.Context, client *domain.Client) error {
	return m.Called(ctx, client).Error(0)
}
func (m *MockCRMClientRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

type MockCRMSubRepo struct {
	mock.Mock
}

func (m *MockCRMSubRepo) Create(ctx context.Context, sub *domain.ClientSubscription) error {
	return m.Called(ctx, sub).Error(0)
}
func (m *MockCRMSubRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ClientSubscription, error) {
	args := m.Called(ctx, id)
	if s := args.Get(0); s != nil {
		return s.(*domain.ClientSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockCRMSubRepo) GetByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.ClientSubscription, error) {
	args := m.Called(ctx, clientID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.ClientSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockCRMSubRepo) Update(ctx context.Context, sub *domain.ClientSubscription) error {
	return m.Called(ctx, sub).Error(0)
}
func (m *MockCRMSubRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func TestClientHandler_ClientsAndSubscriptions(t *testing.T) {
	clientRepo := new(MockCRMClientRepo)
	subRepo := new(MockCRMSubRepo)
	crmSvc := crm.NewService(clientRepo, subRepo)
	tokenMgr := new(MockTokenManager)

	teacherID := uuid.New()
	teacherClaims := &security.UserClaims{
		UserID: teacherID,
		Role:   domain.RoleTeacher,
	}
	tokenMgr.On("ValidateAccessToken", "teacher_token").Return(teacherClaims, nil)

	handler := httpserver.NewAPIHandler(nil, nil, crmSvc, nil, tokenMgr, "v1")

	t.Run("CreateClient success", func(t *testing.T) {
		clientRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.Client) bool {
			return c.TeacherID == teacherID && c.Name == "Мария" && c.BaseRate == 2000
		})).Return(nil).Once()

		reqBody := `{"name":"Мария","base_rate":2000,"school_percent_tag":15}`
		req := httptest.NewRequest(http.MethodPost, "/clients", bytes.NewBufferString(reqBody))
		req.Header.Set("Authorization", "Bearer teacher_token")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.CreateClient(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var resp generated.ClientResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Мария", resp.Name)
		assert.Equal(t, float32(2000), resp.BaseRate)
		assert.Equal(t, 15, resp.SchoolPercentTag)
	})

	t.Run("ListClients success", func(t *testing.T) {
		clients := []*domain.Client{
			{ID: uuid.New(), TeacherID: teacherID, Name: "Мария", BaseRate: 2000, SchoolPercentTag: 15, CreatedAt: time.Now()},
		}
		clientRepo.On("ListByTeacherID", mock.Anything, teacherID).Return(clients, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/clients", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rec := httptest.NewRecorder()

		handler.ListClients(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp []generated.ClientResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "Мария", resp[0].Name)
	})

	t.Run("CreateSubscription success", func(t *testing.T) {
		clientID := uuid.New()
		clientRepo.On("GetByID", mock.Anything, clientID).Return(&domain.Client{
			ID:        clientID,
			TeacherID: teacherID,
		}, nil).Once()
		subRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.ClientSubscription) bool {
			return s.ClientID == clientID && s.Type == domain.SubscriptionTypeLessons && s.Balance == 10
		})).Return(nil).Once()

		reqBody := `{"type":"lessons","balance":10}`
		req := httptest.NewRequest(http.MethodPost, "/clients/"+clientID.String()+"/subscriptions", bytes.NewBufferString(reqBody))
		req.Header.Set("Authorization", "Bearer teacher_token")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.CreateSubscription(rec, req, clientID)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var resp generated.SubscriptionResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "lessons", resp.Type)
		assert.Equal(t, float32(10), resp.Balance)
	})
}
