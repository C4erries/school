package analytics_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/application/analytics"
	"github.com/C4erries/school/backend/internal/application/crm"
	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
)

type MockLessonRepo struct {
	mock.Mock
}

func (m *MockLessonRepo) List(ctx context.Context, filter schedule.LessonFilter) ([]*domain.Lesson, error) {
	args := m.Called(ctx, filter)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Lesson), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockClientRepo struct {
	mock.Mock
}

func (m *MockClientRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	args := m.Called(ctx, id)
	if c := args.Get(0); c != nil {
		return c.(*domain.Client), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockClientRepo) ListByTeacherID(ctx context.Context, teacherID uuid.UUID, filter ...crm.ClientFilter) ([]*domain.Client, error) {
	args := m.Called(ctx, teacherID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Client), args.Error(1)
	}
	return nil, args.Error(1)
}

func setupTestServices() (*analytics.Service, *MockLessonRepo, *MockClientRepo, uuid.UUID, *domain.Client, *domain.Client) {
	lessonRepo := new(MockLessonRepo)
	clientRepo := new(MockClientRepo)
	svc := analytics.NewService(lessonRepo, clientRepo)

	teacherID := uuid.New()
	client1ID := uuid.New()
	client2ID := uuid.New()

	partnerTag := domain.Tag{
		ID:            uuid.New(),
		TeacherID:     teacherID,
		Name:          "Partner School A",
		SchoolPercent: 20, // 20% commission
	}

	r1500 := 1500.0
	r1000 := 1000.0
	client1 := &domain.Client{
		ID:             client1ID,
		TeacherID:      teacherID,
		Name:           "Иван Иванов",
		RateIndividual: 2000.0,
		BaseRate:       2000.0,
		RatePair:       &r1500,
		RateGroup:      &r1000,
		Tags:           []domain.Tag{partnerTag},
	}

	client2 := &domain.Client{
		ID:             client2ID,
		TeacherID:      teacherID,
		Name:           "Анна Смирнова",
		RateIndividual: 1500.0,
		BaseRate:       1500.0,
		Tags:           []domain.Tag{}, // No commission
	}

	return svc, lessonRepo, clientRepo, teacherID, client1, client2
}

func TestAnalyticsService_GetOverview(t *testing.T) {
	svc, lessonRepo, clientRepo, teacherID, client1, client2 := setupTestServices()
	ctx := context.Background()

	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 10, 31, 23, 59, 59, 0, time.UTC)

	// Lessons:
	// 1. client1: 2 hours individual completed -> 2h * 2000 = 4000 gross. 20% comm = 800. Net = 3200.
	// 2. client2: 1.5 hours pair completed -> 1.5h * 1500 (individual/default) = 2250 gross. Net = 2250.
	// 3. client1: 1 hour scheduled (not completed or cancelled) -> ignored in revenue/hours
	// 4. client2: cancelled
	t1Start := time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC)
	t1End := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	l1 := &domain.Lesson{
		ID:        uuid.New(),
		TeacherID: teacherID,
		ClientID:  client1.ID,
		StartTime: t1Start,
		EndTime:   t1End,
		Format:    domain.FormatIndividual,
		Status:    domain.StatusCompleted,
	}

	t2Start := time.Date(2026, 10, 6, 14, 0, 0, 0, time.UTC)
	t2End := time.Date(2026, 10, 6, 15, 30, 0, 0, time.UTC)
	l2 := &domain.Lesson{
		ID:        uuid.New(),
		TeacherID: teacherID,
		ClientID:  client2.ID,
		StartTime: t2Start,
		EndTime:   t2End,
		Format:    domain.FormatIndividual,
		Status:    domain.StatusCompleted,
	}

	t3Start := time.Date(2026, 10, 7, 10, 0, 0, 0, time.UTC)
	t3End := time.Date(2026, 10, 7, 11, 0, 0, 0, time.UTC)
	l3 := &domain.Lesson{
		ID:        uuid.New(),
		TeacherID: teacherID,
		ClientID:  client1.ID,
		StartTime: t3Start,
		EndTime:   t3End,
		Format:    domain.FormatIndividual,
		Status:    domain.StatusScheduled,
	}

	t4Start := time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC)
	t4End := time.Date(2026, 10, 8, 11, 0, 0, 0, time.UTC)
	l4 := &domain.Lesson{
		ID:        uuid.New(),
		TeacherID: teacherID,
		ClientID:  client2.ID,
		StartTime: t4Start,
		EndTime:   t4End,
		Format:    domain.FormatIndividual,
		Status:    domain.StatusCancelled,
	}

	lessonRepo.On("List", ctx, mock.Anything).Return([]*domain.Lesson{l1, l2, l3, l4}, nil)
	clientRepo.On("ListByTeacherID", ctx, teacherID).Return([]*domain.Client{client1, client2}, nil)

	res, err := svc.GetOverview(ctx, teacherID, &from, &to)
	require.NoError(t, err)
	assert.Equal(t, 4, res.TotalLessons)
	assert.Equal(t, 2, res.CompletedLessons)
	assert.Equal(t, 1, res.CancelledLessons)
	// Completion rate = 2 / (2 + 1) * 100 = 66.67%
	assert.InDelta(t, 66.67, res.CompletionRate, 0.1)
	// Completed hours = 2 + 1.5 = 3.5
	assert.InDelta(t, 3.5, res.CompletedHours, 0.01)
	// Gross = 4000 + 2250 = 6250
	assert.InDelta(t, 6250.0, res.GrossRevenue, 0.1)
	// Net = 3200 + 2250 = 5450
	assert.InDelta(t, 5450.0, res.NetIncome, 0.1)
	// Effective rate = 5450 / 3.5 = 1557.14
	assert.InDelta(t, 1557.14, res.EffectiveHourlyRate, 0.1)
}

func TestAnalyticsService_GetDynamics(t *testing.T) {
	svc, lessonRepo, clientRepo, teacherID, client1, client2 := setupTestServices()
	ctx := context.Background()

	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 10, 31, 23, 59, 59, 0, time.UTC)

	t1Start := time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC)
	t1End := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	l1 := &domain.Lesson{
		ID:        uuid.New(),
		TeacherID: teacherID,
		ClientID:  client1.ID,
		StartTime: t1Start,
		EndTime:   t1End,
		Format:    domain.FormatIndividual,
		Status:    domain.StatusCompleted,
	}

	lessonRepo.On("List", ctx, mock.Anything).Return([]*domain.Lesson{l1}, nil)
	clientRepo.On("ListByTeacherID", ctx, teacherID).Return([]*domain.Client{client1, client2}, nil)

	points, err := svc.GetDynamics(ctx, teacherID, "week", &from, &to)
	require.NoError(t, err)
	assert.NotEmpty(t, points)

	// Check points by month
	monthPoints, err := svc.GetDynamics(ctx, teacherID, "month", &from, &to)
	require.NoError(t, err)
	assert.NotEmpty(t, monthPoints)
	assert.Equal(t, 1, len(monthPoints))
	assert.Equal(t, 1, monthPoints[0].CompletedCount)
	assert.InDelta(t, 2.0, monthPoints[0].CompletedHours, 0.01)
	assert.InDelta(t, 3200.0, monthPoints[0].NetIncome, 0.01)
}

func TestAnalyticsService_GetFormats(t *testing.T) {
	svc, lessonRepo, clientRepo, teacherID, client1, client2 := setupTestServices()
	ctx := context.Background()

	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 10, 31, 23, 59, 59, 0, time.UTC)

	// 2h individual: 2h * 2000 - 20% = 3200
	l1 := &domain.Lesson{
		ID:        uuid.New(),
		TeacherID: teacherID,
		ClientID:  client1.ID,
		StartTime: time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC),
		Format:    domain.FormatIndividual,
		Status:    domain.StatusCompleted,
	}
	// 1h pair: 1h * 1500 - 20% = 1200
	l2 := &domain.Lesson{
		ID:        uuid.New(),
		TeacherID: teacherID,
		ClientID:  client1.ID,
		StartTime: time.Date(2026, 10, 6, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 10, 6, 11, 0, 0, 0, time.UTC),
		Format:    domain.FormatPair,
		Status:    domain.StatusCompleted,
	}

	lessonRepo.On("List", ctx, mock.Anything).Return([]*domain.Lesson{l1, l2}, nil)
	clientRepo.On("ListByTeacherID", ctx, teacherID).Return([]*domain.Client{client1, client2}, nil)

	formats, err := svc.GetFormats(ctx, teacherID, &from, &to)
	require.NoError(t, err)
	assert.Len(t, formats, 3)

	indiv := formats[0]
	assert.Equal(t, domain.FormatIndividual, indiv.Format)
	assert.InDelta(t, 2.0, indiv.CompletedHours, 0.01)
	assert.InDelta(t, 3200.0, indiv.NetIncome, 0.01)
	assert.InDelta(t, 66.7, indiv.HoursSharePercent, 0.1)

	pair := formats[1]
	assert.Equal(t, domain.FormatPair, pair.Format)
	assert.InDelta(t, 1.0, pair.CompletedHours, 0.01)
	assert.InDelta(t, 1200.0, pair.NetIncome, 0.01)
	assert.InDelta(t, 33.3, pair.HoursSharePercent, 0.1)

	group := formats[2]
	assert.Equal(t, domain.FormatGroup, group.Format)
	assert.InDelta(t, 0.0, group.CompletedHours, 0.01)
}

func TestAnalyticsService_GetClients(t *testing.T) {
	svc, lessonRepo, clientRepo, teacherID, client1, client2 := setupTestServices()
	ctx := context.Background()

	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 10, 31, 23, 59, 59, 0, time.UTC)

	// client1: 2h completed (3200 net), 0 cancelled
	l1 := &domain.Lesson{
		ID:        uuid.New(),
		TeacherID: teacherID,
		ClientID:  client1.ID,
		StartTime: time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC),
		Format:    domain.FormatIndividual,
		Status:    domain.StatusCompleted,
	}
	// client2: 1h completed (1500 net), 2 cancelled
	l2 := &domain.Lesson{
		ID:        uuid.New(),
		TeacherID: teacherID,
		ClientID:  client2.ID,
		StartTime: time.Date(2026, 10, 6, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 10, 6, 11, 0, 0, 0, time.UTC),
		Format:    domain.FormatIndividual,
		Status:    domain.StatusCompleted,
	}
	l3 := &domain.Lesson{
		ID:        uuid.New(),
		TeacherID: teacherID,
		ClientID:  client2.ID,
		StartTime: time.Date(2026, 10, 7, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 10, 7, 11, 0, 0, 0, time.UTC),
		Format:    domain.FormatIndividual,
		Status:    domain.StatusCancelled,
	}
	l4 := &domain.Lesson{
		ID:        uuid.New(),
		TeacherID: teacherID,
		ClientID:  client2.ID,
		StartTime: time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 10, 8, 11, 0, 0, 0, time.UTC),
		Format:    domain.FormatIndividual,
		Status:    domain.StatusCancelled,
	}

	lessonRepo.On("List", ctx, mock.Anything).Return([]*domain.Lesson{l1, l2, l3, l4}, nil)
	clientRepo.On("ListByTeacherID", ctx, teacherID).Return([]*domain.Client{client1, client2}, nil)

	// Sort by hours
	stats, err := svc.GetClients(ctx, teacherID, "hours", 10, &from, &to)
	require.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, client1.ID, stats[0].ClientID)
	assert.Equal(t, "Иван Иванов", stats[0].ClientName)
	assert.InDelta(t, 2.0, stats[0].CompletedHours, 0.01)
	assert.InDelta(t, 100.0, stats[0].AttendanceRate, 0.1)

	// Sort by cancellations
	statsCanc, err := svc.GetClients(ctx, teacherID, "cancellations", 10, &from, &to)
	require.NoError(t, err)
	assert.Equal(t, client2.ID, statsCanc[0].ClientID)
	assert.Equal(t, 2, statsCanc[0].CancelledCount)
	// Attendance rate for client2: 1 / (1 + 2) = 33.3%
	assert.InDelta(t, 33.3, statsCanc[0].AttendanceRate, 0.1)
}
