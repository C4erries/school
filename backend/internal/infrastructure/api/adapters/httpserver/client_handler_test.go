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
func (m *MockCRMClientRepo) ListByTeacherID(ctx context.Context, teacherID uuid.UUID, filter ...crm.ClientFilter) ([]*domain.Client, error) {
	if len(filter) > 0 {
		args := m.Called(ctx, teacherID, filter[0])
		if list := args.Get(0); list != nil {
			return list.([]*domain.Client), args.Error(1)
		}
		return nil, args.Error(1)
	}
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

type MockCRMBalanceAdjustmentRepo struct {
	mock.Mock
}

func (m *MockCRMBalanceAdjustmentRepo) Create(ctx context.Context, adj *domain.ClientBalanceAdjustment) error {
	return m.Called(ctx, adj).Error(0)
}

func (m *MockCRMBalanceAdjustmentRepo) ListByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.ClientBalanceAdjustment, error) {
	args := m.Called(ctx, clientID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.ClientBalanceAdjustment), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestClientHandler_ClientsAndSubscriptions(t *testing.T) {
	clientRepo := new(MockCRMClientRepo)
	subRepo := new(MockCRMSubRepo)
	adjRepo := new(MockCRMBalanceAdjustmentRepo)
	crmSvc := crm.NewService(clientRepo, subRepo, adjRepo)
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
			return c.TeacherID == teacherID && c.Name == "Мария" && c.RateIndividual == 2000
		})).Return(nil).Once()

		reqBody := `{"name":"Мария","rate_individual":2000,"school_percent_tag":15}`
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
		assert.Equal(t, float32(2000), resp.RateIndividual)
	})

	t.Run("ListClients success", func(t *testing.T) {
		clients := []*domain.Client{
			{
				ID:             uuid.New(),
				TeacherID:      teacherID,
				Name:           "Мария",
				RateIndividual: 2000,
				CreatedAt:      time.Now(),
				Balances: domain.ClientBalances{
					IndividualHours: 8.5,
					PairHours:       4.0,
					TotalHours:      12.5,
				},
			},
		}
		clientRepo.On("ListByTeacherID", mock.Anything, teacherID).Return(clients, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/clients", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rec := httptest.NewRecorder()

		handler.ListClients(rec, req, generated.ListClientsParams{})

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp []generated.ClientResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "Мария", resp[0].Name)
		assert.Equal(t, float32(8.5), resp[0].Balances.IndividualHours)
		assert.Equal(t, float32(4.0), resp[0].Balances.PairHours)
		assert.Equal(t, float32(12.5), resp[0].Balances.TotalHours)
	})

	t.Run("ListClients with filter", func(t *testing.T) {
		isArchived := true
		search := "Иван"
		expectedFilter := crm.ClientFilter{
			IsArchived: &isArchived,
			Search:     &search,
		}
		clients := []*domain.Client{
			{
				ID:             uuid.New(),
				TeacherID:      teacherID,
				Name:           "Иван",
				RateIndividual: 1500,
				IsArchived:     true,
				CreatedAt:      time.Now(),
			},
		}
		clientRepo.On("ListByTeacherID", mock.Anything, teacherID, expectedFilter).Return(clients, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/clients?search=Иван&is_archived=true", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rec := httptest.NewRecorder()

		handler.ListClients(rec, req, generated.ListClientsParams{
			IsArchived: &isArchived,
			Search:     &search,
		})

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp []generated.ClientResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "Иван", resp[0].Name)
		assert.True(t, resp[0].IsArchived)
	})

	t.Run("ArchiveClient success", func(t *testing.T) {
		clientID := uuid.New()
		client := &domain.Client{
			ID:             clientID,
			TeacherID:      teacherID,
			Name:           "Мария",
			RateIndividual: 2000,
			IsArchived:     false,
			CreatedAt:      time.Now(),
		}
		archivedClient := &domain.Client{
			ID:             clientID,
			TeacherID:      teacherID,
			Name:           "Мария",
			RateIndividual: 2000,
			IsArchived:     true,
			CreatedAt:      time.Now(),
		}
		clientRepo.On("GetByID", mock.Anything, clientID).Return(client, nil).Once()
		clientRepo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Client) bool {
			return c.ID == clientID && c.IsArchived
		})).Return(nil).Once()
		clientRepo.On("GetByID", mock.Anything, clientID).Return(archivedClient, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/clients/"+clientID.String()+"/archive", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rec := httptest.NewRecorder()

		handler.ArchiveClient(rec, req, clientID)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp generated.ClientResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.True(t, resp.IsArchived)
	})

	t.Run("UnarchiveClient success", func(t *testing.T) {
		clientID := uuid.New()
		client := &domain.Client{
			ID:             clientID,
			TeacherID:      teacherID,
			Name:           "Мария",
			RateIndividual: 2000,
			IsArchived:     true,
			CreatedAt:      time.Now(),
		}
		unarchivedClient := &domain.Client{
			ID:             clientID,
			TeacherID:      teacherID,
			Name:           "Мария",
			RateIndividual: 2000,
			IsArchived:     false,
			CreatedAt:      time.Now(),
		}
		clientRepo.On("GetByID", mock.Anything, clientID).Return(client, nil).Once()
		clientRepo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Client) bool {
			return c.ID == clientID && !c.IsArchived
		})).Return(nil).Once()
		clientRepo.On("GetByID", mock.Anything, clientID).Return(unarchivedClient, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/clients/"+clientID.String()+"/unarchive", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rec := httptest.NewRecorder()

		handler.UnarchiveClient(rec, req, clientID)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp generated.ClientResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.False(t, resp.IsArchived)
	})

	t.Run("AdjustClientBalance success", func(t *testing.T) {
		clientID := uuid.New()
		client := &domain.Client{
			ID:             clientID,
			TeacherID:      teacherID,
			Name:           "Мария",
			RateIndividual: 2000,
			CreatedAt:      time.Now(),
		}
		updatedClient := &domain.Client{
			ID:             clientID,
			TeacherID:      teacherID,
			Name:           "Мария",
			RateIndividual: 2000,
			Balances: domain.ClientBalances{
				IndividualHours: 7.0,
			},
			CreatedAt: time.Now(),
		}
		sub := &domain.ClientSubscription{
			ID:        uuid.New(),
			ClientID:  clientID,
			Format:    domain.SubscriptionFormatIndividual,
			Balance:   5.0,
			CreatedAt: time.Now(),
		}

		clientRepo.On("GetByID", mock.Anything, clientID).Return(client, nil).Once()
		subRepo.On("GetByClientID", mock.Anything, clientID).Return([]*domain.ClientSubscription{sub}, nil).Once()
		subRepo.On("Update", mock.Anything, mock.MatchedBy(func(s *domain.ClientSubscription) bool {
			return s.ID == sub.ID && s.Balance == 7.0
		})).Return(nil).Once()
		adjRepo.On("Create", mock.Anything, mock.MatchedBy(func(a *domain.ClientBalanceAdjustment) bool {
			return a.ClientID == clientID &&
				a.TeacherID == teacherID &&
				a.Format == domain.SubscriptionFormatIndividual &&
				a.DeltaHours == 2.0 &&
				a.Reason == "Ручная корректировка после болезни"
		})).Return(nil).Once()
		clientRepo.On("GetByID", mock.Anything, clientID).Return(updatedClient, nil).Once()

		reqBody := `{"delta_hours":2.0,"format":"individual","reason":"Ручная корректировка после болезни"}`
		req := httptest.NewRequest(http.MethodPost, "/clients/"+clientID.String()+"/adjust-balance", bytes.NewBufferString(reqBody))
		req.Header.Set("Authorization", "Bearer teacher_token")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.AdjustClientBalance(rec, req, clientID)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp generated.ClientResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, float32(7.0), resp.Balances.IndividualHours)
	})

	t.Run("CreateSubscription success", func(t *testing.T) {
		clientID := uuid.New()
		clientRepo.On("GetByID", mock.Anything, clientID).Return(&domain.Client{
			ID:        clientID,
			TeacherID: teacherID,
		}, nil).Once()
		subRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.ClientSubscription) bool {
			return s.ClientID == clientID && s.Format == domain.SubscriptionFormatIndividual && s.Balance == 10
		})).Return(nil).Once()

		reqBody := `{"format":"individual","balance":10}`
		req := httptest.NewRequest(http.MethodPost, "/clients/"+clientID.String()+"/subscriptions", bytes.NewBufferString(reqBody))
		req.Header.Set("Authorization", "Bearer teacher_token")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.CreateSubscription(rec, req, clientID)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var resp generated.SubscriptionResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, generated.SubscriptionResponseFormatIndividual, resp.Format)
		assert.Equal(t, float32(10), resp.Balance)
	})
}

