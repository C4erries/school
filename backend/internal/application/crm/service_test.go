package crm_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/application/crm"
	"github.com/C4erries/school/backend/internal/domain"
)

type MockClientRepository struct {
	mock.Mock
}

func (m *MockClientRepository) Create(ctx context.Context, client *domain.Client) error {
	return m.Called(ctx, client).Error(0)
}
func (m *MockClientRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	args := m.Called(ctx, id)
	if c := args.Get(0); c != nil {
		return c.(*domain.Client), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockClientRepository) ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.Client, error) {
	args := m.Called(ctx, teacherID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Client), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockClientRepository) Update(ctx context.Context, client *domain.Client) error {
	return m.Called(ctx, client).Error(0)
}
func (m *MockClientRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) Create(ctx context.Context, sub *domain.ClientSubscription) error {
	return m.Called(ctx, sub).Error(0)
}
func (m *MockSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ClientSubscription, error) {
	args := m.Called(ctx, id)
	if s := args.Get(0); s != nil {
		return s.(*domain.ClientSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockSubscriptionRepository) GetByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.ClientSubscription, error) {
	args := m.Called(ctx, clientID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.ClientSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockSubscriptionRepository) Update(ctx context.Context, sub *domain.ClientSubscription) error {
	return m.Called(ctx, sub).Error(0)
}
func (m *MockSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func TestCRMService_Clients(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(MockClientRepository)
	subRepo := new(MockSubscriptionRepository)
	svc := crm.NewService(clientRepo, subRepo)

	teacherID := uuid.New()
	phone := "+79991234567"

	t.Run("CreateClient", func(t *testing.T) {
		clientRepo.On("Create", ctx, mock.MatchedBy(func(c *domain.Client) bool {
			return c.TeacherID == teacherID && c.Name == "Иван" && c.BaseRate == 1500 && c.SchoolPercentTag == 10
		})).Return(nil).Once()

		c, err := svc.CreateClient(ctx, teacherID, "Иван", &phone, 1500, 10)
		require.NoError(t, err)
		assert.Equal(t, "Иван", c.Name)
		assert.Equal(t, 1500.0, c.BaseRate)
		assert.Equal(t, 10, c.SchoolPercentTag)
		clientRepo.AssertExpectations(t)
	})

	t.Run("ListClients", func(t *testing.T) {
		expected := []*domain.Client{
			{ID: uuid.New(), TeacherID: teacherID, Name: "Иван"},
		}
		clientRepo.On("ListByTeacherID", ctx, teacherID).Return(expected, nil).Once()

		res, err := svc.ListClients(ctx, teacherID)
		require.NoError(t, err)
		assert.Equal(t, expected, res)
		clientRepo.AssertExpectations(t)
	})

	t.Run("GetClient", func(t *testing.T) {
		cid := uuid.New()
		expected := &domain.Client{ID: cid, TeacherID: teacherID, Name: "Иван"}
		clientRepo.On("GetByID", ctx, cid).Return(expected, nil).Once()

		res, err := svc.GetClient(ctx, cid)
		require.NoError(t, err)
		assert.Equal(t, expected, res)
		clientRepo.AssertExpectations(t)
	})
}

func TestCRMService_Subscriptions(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(MockClientRepository)
	subRepo := new(MockSubscriptionRepository)
	svc := crm.NewService(clientRepo, subRepo)

	clientID := uuid.New()

	t.Run("CreateSubscription", func(t *testing.T) {
		subRepo.On("Create", ctx, mock.MatchedBy(func(s *domain.ClientSubscription) bool {
			return s.ClientID == clientID && s.Format == domain.SubscriptionFormatIndividual && s.Balance == 8
		})).Return(nil).Once()

		s, err := svc.CreateSubscription(ctx, clientID, domain.SubscriptionFormatIndividual, 8)
		require.NoError(t, err)
		assert.Equal(t, domain.SubscriptionFormatIndividual, s.Format)
		assert.Equal(t, 8.0, s.Balance)
		subRepo.AssertExpectations(t)
	})

	t.Run("ListSubscriptions", func(t *testing.T) {
		expected := []*domain.ClientSubscription{
			{ID: uuid.New(), ClientID: clientID, Format: domain.SubscriptionFormatIndividual, Balance: 8, CreatedAt: time.Now()},
		}
		subRepo.On("GetByClientID", ctx, clientID).Return(expected, nil).Once()

		res, err := svc.ListSubscriptions(ctx, clientID)
		require.NoError(t, err)
		assert.Equal(t, expected, res)
		subRepo.AssertExpectations(t)
	})
}

type MockTagRepository struct {
	mock.Mock
}

func (m *MockTagRepository) Create(ctx context.Context, tag *domain.Tag) error {
	return m.Called(ctx, tag).Error(0)
}
func (m *MockTagRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tag, error) {
	args := m.Called(ctx, id)
	if t := args.Get(0); t != nil {
		return t.(*domain.Tag), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockTagRepository) ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.Tag, error) {
	args := m.Called(ctx, teacherID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Tag), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockTagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockTagRepository) AssignToClient(ctx context.Context, clientID, tagID uuid.UUID) error {
	return m.Called(ctx, clientID, tagID).Error(0)
}
func (m *MockTagRepository) RemoveFromClient(ctx context.Context, clientID, tagID uuid.UUID) error {
	return m.Called(ctx, clientID, tagID).Error(0)
}
func (m *MockTagRepository) SetClientTags(ctx context.Context, clientID uuid.UUID, tagIDs []uuid.UUID) error {
	return m.Called(ctx, clientID, tagIDs).Error(0)
}

func TestCRMService_TagsAndRates(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(MockClientRepository)
	subRepo := new(MockSubscriptionRepository)
	tagRepo := new(MockTagRepository)
	svc := crm.NewService(clientRepo, subRepo, tagRepo)

	teacherID := uuid.New()

	t.Run("CreateTag success", func(t *testing.T) {
		tagRepo.On("Create", ctx, mock.MatchedBy(func(tag *domain.Tag) bool {
			return tag.TeacherID == teacherID && tag.Name == "Школа №12" && tag.SchoolPercent == 30 && tag.Color == "indigo"
		})).Return(nil).Once()

		tag, err := svc.CreateTag(ctx, teacherID, "Школа №12", 30, "indigo")
		require.NoError(t, err)
		assert.Equal(t, "Школа №12", tag.Name)
		assert.Equal(t, 30, tag.SchoolPercent)
		tagRepo.AssertExpectations(t)
	})

	t.Run("CreateClientWithRates", func(t *testing.T) {
		ratePair := 1000.0
		rateGroup := 800.0
		clientRepo.On("Create", ctx, mock.MatchedBy(func(c *domain.Client) bool {
			return c.TeacherID == teacherID && c.Name == "Алексей" && c.RateIndividual == 1500 && *c.RatePair == 1000
		})).Return(nil).Once()

		c, err := svc.CreateClientWithRates(ctx, crm.CreateClientInput{
			TeacherID:      teacherID,
			Name:           "Алексей",
			RateIndividual: 1500,
			RatePair:       &ratePair,
			RateGroup:      &rateGroup,
		})
		require.NoError(t, err)
		assert.Equal(t, "Алексей", c.Name)
		assert.Equal(t, 1500.0, c.RateIndividual)
		assert.Equal(t, 1000.0, *c.RatePair)
		clientRepo.AssertExpectations(t)
	})

	t.Run("AssignTagToClient", func(t *testing.T) {
		clientID := uuid.New()
		tagID := uuid.New()

		clientRepo.On("GetByID", ctx, clientID).Return(&domain.Client{
			ID:        clientID,
			TeacherID: teacherID,
		}, nil).Twice()
		tagRepo.On("GetByID", ctx, tagID).Return(&domain.Tag{
			ID:        tagID,
			TeacherID: teacherID,
		}, nil).Once()
		tagRepo.On("AssignToClient", ctx, clientID, tagID).Return(nil).Once()

		_, err := svc.AssignTagToClient(ctx, clientID, tagID, teacherID, domain.RoleTeacher)
		require.NoError(t, err)
		tagRepo.AssertExpectations(t)
		clientRepo.AssertExpectations(t)
	})
}
