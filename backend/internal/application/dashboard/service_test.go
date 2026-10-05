package dashboard_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/application/dashboard"
	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
)

type MockClientRepository struct {
	mock.Mock
}

func (m *MockClientRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	args := m.Called(ctx, id)
	if c := args.Get(0); c != nil {
		return c.(*domain.Client), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockLessonRepository struct {
	mock.Mock
}

func (m *MockLessonRepository) List(ctx context.Context, filter schedule.LessonFilter) ([]*domain.Lesson, error) {
	args := m.Called(ctx, filter)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Lesson), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestDashboardService_GetMetrics(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(MockClientRepository)
	lessonRepo := new(MockLessonRepository)
	svc := dashboard.NewService(lessonRepo, clientRepo)

	teacherID := uuid.New()
	client1ID := uuid.New()
	client2ID := uuid.New()

	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 10, 31, 23, 59, 59, 0, time.UTC)

	client1 := &domain.Client{
		ID:               client1ID,
		TeacherID:        teacherID,
		Name:             "Клиент 1",
		BaseRate:         2000,
		SchoolPercentTag: 20, // 20% commission
	}
	client2 := &domain.Client{
		ID:               client2ID,
		TeacherID:        teacherID,
		Name:             "Клиент 2",
		BaseRate:         1000,
		SchoolPercentTag: 0, // 0% commission
	}

	lessons := []*domain.Lesson{
		// 1 hour lesson with client1: completed (2000 gross, net = 2000 - 400 = 1600)
		{
			ID:        uuid.New(),
			TeacherID: teacherID,
			ClientID:  client1ID,
			StartTime: from.Add(2 * time.Hour),
			EndTime:   from.Add(3 * time.Hour),
			Status:    domain.StatusCompleted,
		},
		// 2 hours lesson with client2: scheduled (2000 gross, net = 2000 - 0 = 2000)
		{
			ID:        uuid.New(),
			TeacherID: teacherID,
			ClientID:  client2ID,
			StartTime: from.Add(4 * time.Hour),
			EndTime:   from.Add(6 * time.Hour),
			Status:    domain.StatusScheduled,
		},
		// cancelled lesson (should be ignored)
		{
			ID:        uuid.New(),
			TeacherID: teacherID,
			ClientID:  client1ID,
			StartTime: from.Add(7 * time.Hour),
			EndTime:   from.Add(8 * time.Hour),
			Status:    domain.StatusCancelled,
		},
	}

	lessonRepo.On("List", ctx, schedule.LessonFilter{
		TeacherID: &teacherID,
		From:      &from,
		To:        &to,
	}).Return(lessons, nil)

	clientRepo.On("GetByID", ctx, client1ID).Return(client1, nil)
	clientRepo.On("GetByID", ctx, client2ID).Return(client2, nil)

	metrics, err := svc.GetMetrics(ctx, teacherID, from, to)
	require.NoError(t, err)

	// Gross: 1h*2000 + 2h*1000 = 4000
	assert.Equal(t, 4000.0, metrics.GrossPotentialRevenue)
	// Net: 1600 + 2000 = 3600
	assert.Equal(t, 3600.0, metrics.NetIncome)
	// Average rate: (2000 + 1000) / 2 clients = 1500
	assert.Equal(t, 1500.0, metrics.AverageRate)
}
