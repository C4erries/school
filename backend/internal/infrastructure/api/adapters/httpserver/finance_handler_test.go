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

	"github.com/C4erries/school/backend/internal/application/finance"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/security"
)

type MockFinancePaymentRepo struct {
	mock.Mock
}

func (m *MockFinancePaymentRepo) Create(ctx context.Context, p *domain.Payment) error {
	return m.Called(ctx, p).Error(0)
}
func (m *MockFinancePaymentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	args := m.Called(ctx, id)
	if p := args.Get(0); p != nil {
		return p.(*domain.Payment), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockFinancePaymentRepo) List(ctx context.Context, filter finance.PaymentFilter) ([]*domain.Payment, error) {
	args := m.Called(ctx, filter)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Payment), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockFinancePaymentRepo) SumAmountByPeriod(ctx context.Context, teacherID uuid.UUID, from, to time.Time) (float64, error) {
	args := m.Called(ctx, teacherID, from, to)
	return args.Get(0).(float64), args.Error(1)
}

type MockFinancePayoutRepo struct {
	mock.Mock
}

func (m *MockFinancePayoutRepo) Create(ctx context.Context, p *domain.PartnerPayout) error {
	return m.Called(ctx, p).Error(0)
}
func (m *MockFinancePayoutRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PartnerPayout, error) {
	args := m.Called(ctx, id)
	if p := args.Get(0); p != nil {
		return p.(*domain.PartnerPayout), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockFinancePayoutRepo) GetByTagAndPeriod(ctx context.Context, teacherID, tagID uuid.UUID, periodMonth string) (*domain.PartnerPayout, error) {
	args := m.Called(ctx, teacherID, tagID, periodMonth)
	if p := args.Get(0); p != nil {
		return p.(*domain.PartnerPayout), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockFinancePayoutRepo) ListByPeriod(ctx context.Context, teacherID uuid.UUID, periodMonth string) ([]*domain.PartnerPayout, error) {
	args := m.Called(ctx, teacherID, periodMonth)
	if list := args.Get(0); list != nil {
		return list.([]*domain.PartnerPayout), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockFinancePayoutRepo) ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.PartnerPayout, error) {
	args := m.Called(ctx, teacherID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.PartnerPayout), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockCRMTagRepo struct {
	mock.Mock
}

func (m *MockCRMTagRepo) Create(ctx context.Context, tag *domain.Tag) error {
	return m.Called(ctx, tag).Error(0)
}
func (m *MockCRMTagRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tag, error) {
	args := m.Called(ctx, id)
	if t := args.Get(0); t != nil {
		return t.(*domain.Tag), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockCRMTagRepo) ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.Tag, error) {
	args := m.Called(ctx, teacherID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Tag), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockCRMTagRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockCRMTagRepo) AssignToClient(ctx context.Context, clientID, tagID uuid.UUID) error {
	return m.Called(ctx, clientID, tagID).Error(0)
}
func (m *MockCRMTagRepo) RemoveFromClient(ctx context.Context, clientID, tagID uuid.UUID) error {
	return m.Called(ctx, clientID, tagID).Error(0)
}
func (m *MockCRMTagRepo) SetClientTags(ctx context.Context, clientID uuid.UUID, tagIDs []uuid.UUID) error {
	return m.Called(ctx, clientID, tagIDs).Error(0)
}

func TestFinanceHandler_Endpoints(t *testing.T) {
	teacherID := uuid.New()
	clientID := uuid.New()
	tagID := uuid.New()

	paymentRepo := new(MockFinancePaymentRepo)
	payoutRepo := new(MockFinancePayoutRepo)
	clientRepo := new(MockCRMClientRepo)
	subRepo := new(MockCRMSubRepo)
	lessonRepo := new(MockLessonRepo)
	tagRepo := new(MockCRMTagRepo)
	tokenMgr := new(MockTokenManager)

	tokenMgr.On("ValidateAccessToken", "teacher_token").Return(&security.UserClaims{
		UserID: teacherID,
		Role:   domain.RoleTeacher,
	}, nil)

	financeSvc := finance.NewService(paymentRepo, payoutRepo, clientRepo, subRepo, lessonRepo, tagRepo)
	handler := httpserver.NewAPIHandler(nil, nil, nil, nil, tokenMgr, "v1", financeSvc)
	mux := httpserver.BuildMux(handler, nil)

	client := &domain.Client{
		ID:             clientID,
		TeacherID:      teacherID,
		Name:           "Тестовый Ученик",
		RateIndividual: 1500,
	}

	t.Run("POST /finance/payments - success", func(t *testing.T) {
		clientRepo.On("GetByID", mock.Anything, clientID).Return(client, nil).Once()
		paymentRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *domain.Payment) bool {
			return p.Amount == 4500 && p.Hours == 3
		})).Return(nil).Once()
		subRepo.On("GetByClientID", mock.Anything, clientID).Return([]*domain.ClientSubscription{}, nil).Once()
		subRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.ClientSubscription) bool {
			return s.ClientID == clientID && s.Balance == 3
		})).Return(nil).Once()

		reqBody := generated.CreatePaymentRequest{
			ClientId: clientID,
			Amount:   4500,
			Hours:    3,
			Format:   generated.CreatePaymentRequestFormatIndividual,
		}
		data, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/finance/payments", bytes.NewReader(data))
		req.Header.Set("Authorization", "Bearer teacher_token")
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		require.Equal(t, http.StatusCreated, rr.Code)
		var resp generated.PaymentResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, float32(4500), resp.Amount)
		assert.Equal(t, float32(3), resp.Hours)
		assert.Equal(t, "Тестовый Ученик", *resp.ClientName)
	})

	t.Run("GET /finance/payments - success", func(t *testing.T) {
		payments := []*domain.Payment{
			{
				ID:            uuid.New(),
				TeacherID:     teacherID,
				ClientID:      clientID,
				Amount:        4500,
				Hours:         3,
				Format:        domain.SubscriptionFormatIndividual,
				PaymentMethod: domain.PaymentMethodTransfer,
				PaidAt:        time.Now().UTC(),
			},
		}

		paymentRepo.On("List", mock.Anything, mock.Anything).Return(payments, nil).Once()
		clientRepo.On("ListByTeacherID", mock.Anything, teacherID).Return([]*domain.Client{client}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/finance/payments", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp []generated.PaymentResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		require.Len(t, resp, 1)
		assert.Equal(t, float32(4500), resp[0].Amount)
	})

	t.Run("GET /finance/summary - success", func(t *testing.T) {
		paymentRepo.On("SumAmountByPeriod", mock.Anything, teacherID, mock.Anything, mock.Anything).Return(15000.0, nil).Once()
		clientRepo.On("ListByTeacherID", mock.Anything, teacherID).Return([]*domain.Client{client}, nil).Once()
		lessonRepo.On("List", mock.Anything, mock.Anything).Return([]*domain.Lesson{}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/finance/summary?month=2026-10", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp generated.FinanceSummaryResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "2026-10", resp.Month)
		assert.Equal(t, float32(15000), resp.TotalPayments)
	})

	t.Run("GET /finance/partner-settlements - success", func(t *testing.T) {
		tag := &domain.Tag{
			ID:            tagID,
			TeacherID:     teacherID,
			Name:          "Партнерская школа",
			SchoolPercent: 20,
		}
		tagRepo.On("ListByTeacherID", mock.Anything, teacherID).Return([]*domain.Tag{tag}, nil).Once()
		payoutRepo.On("ListByPeriod", mock.Anything, teacherID, "2026-10").Return([]*domain.PartnerPayout{}, nil).Once()
		lessonRepo.On("List", mock.Anything, mock.Anything).Return([]*domain.Lesson{}, nil).Once()
		clientRepo.On("ListByTeacherID", mock.Anything, teacherID).Return([]*domain.Client{}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/finance/partner-settlements?month=2026-10", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp []generated.PartnerSettlementResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		require.Len(t, resp, 1)
		assert.Equal(t, "Партнерская школа", resp[0].TagName)
		assert.False(t, resp[0].IsPaid)
	})

	t.Run("POST /finance/partner-payouts - success", func(t *testing.T) {
		tag := &domain.Tag{
			ID:            tagID,
			TeacherID:     teacherID,
			Name:          "Партнерская школа",
			SchoolPercent: 20,
		}
		tagRepo.On("GetByID", mock.Anything, tagID).Return(tag, nil).Once()
		payoutRepo.On("GetByTagAndPeriod", mock.Anything, teacherID, tagID, "2026-10").Return(nil, domain.ErrPartnerPayoutNotFound).Once()
		payoutRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		reqBody := generated.CreatePartnerPayoutRequest{
			TagId:            tagID,
			PeriodMonth:      "2026-10",
			GrossAmount:      20000,
			CommissionAmount: 4000,
		}
		data, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/finance/partner-payouts", bytes.NewReader(data))
		req.Header.Set("Authorization", "Bearer teacher_token")
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		require.Equal(t, http.StatusCreated, rr.Code)
		var resp generated.PartnerPayoutResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, float32(4000), resp.CommissionAmount)
	})

	t.Run("POST /finance/partner-payouts - duplicate conflict 409", func(t *testing.T) {
		tag := &domain.Tag{
			ID:            tagID,
			TeacherID:     teacherID,
			Name:          "Партнерская школа",
			SchoolPercent: 20,
		}
		existing := &domain.PartnerPayout{
			ID:          uuid.New(),
			TeacherID:   teacherID,
			TagID:       tagID,
			PeriodMonth: "2026-10",
		}
		tagRepo.On("GetByID", mock.Anything, tagID).Return(tag, nil).Once()
		payoutRepo.On("GetByTagAndPeriod", mock.Anything, teacherID, tagID, "2026-10").Return(existing, nil).Once()

		reqBody := generated.CreatePartnerPayoutRequest{
			TagId:            tagID,
			PeriodMonth:      "2026-10",
			GrossAmount:      20000,
			CommissionAmount: 4000,
		}
		data, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/finance/partner-payouts", bytes.NewReader(data))
		req.Header.Set("Authorization", "Bearer teacher_token")
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		require.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("GET /export/clients - success UTF-8 BOM", func(t *testing.T) {
		clientRepo.On("ListByTeacherID", mock.Anything, teacherID).Return([]*domain.Client{client}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/export/clients", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "text/csv; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.True(t, bytes.HasPrefix(rr.Body.Bytes(), []byte("\xEF\xBB\xBF")))
	})
}
