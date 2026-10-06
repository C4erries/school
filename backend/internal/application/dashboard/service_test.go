package dashboard_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/application/crm"
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

func (m *MockClientRepository) ListByTeacherID(ctx context.Context, teacherID uuid.UUID, filter ...crm.ClientFilter) ([]*domain.Client, error) {
	args := m.Called(ctx, teacherID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Client), args.Error(1)
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

type MockClassroomRepository struct {
	mock.Mock
}

func (m *MockClassroomRepository) List(ctx context.Context) ([]*domain.Classroom, error) {
	args := m.Called(ctx)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Classroom), args.Error(1)
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

func TestDashboardService_GetSummary(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(MockClientRepository)
	lessonRepo := new(MockLessonRepository)
	classroomRepo := new(MockClassroomRepository)
	svc := dashboard.NewService(lessonRepo, clientRepo, classroomRepo)

	teacherID := uuid.New()
	client1ID := uuid.New()
	client2ID := uuid.New()
	classroomID := uuid.New()

	now := time.Now().UTC()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	partnerTag := domain.Tag{
		ID:            uuid.New(),
		TeacherID:     teacherID,
		Name:          "Партнер",
		SchoolPercent: 20,
	}

	client1 := &domain.Client{
		ID:             client1ID,
		TeacherID:      teacherID,
		Name:           "Алексей Иванов",
		RateIndividual: 2000,
		BaseRate:       2000,
		Tags:           []domain.Tag{partnerTag},
		Balances: domain.ClientBalances{
			IndividualHours: -2, // Долг 2h * 2000 = 4000
		},
	}
	client2 := &domain.Client{
		ID:             client2ID,
		TeacherID:      teacherID,
		Name:           "Мария Петрова",
		RateIndividual: 1500,
		BaseRate:       1500,
		Tags:           []domain.Tag{},
		Balances: domain.ClientBalances{
			IndividualHours: 5, // Положительный баланс
		},
	}

	classroom := &domain.Classroom{
		ID:    classroomID,
		Name:  "Аудитория 101",
		Color: "#3B82F6",
	}

	todayLesson := &domain.Lesson{
		ID:          uuid.New(),
		TeacherID:   teacherID,
		ClientID:    client1ID,
		ClassroomID: &classroomID,
		StartTime:   startOfToday.Add(10 * time.Hour),
		EndTime:     startOfToday.Add(11 * time.Hour),
		Format:      domain.FormatIndividual,
		Status:      domain.StatusScheduled,
	}

	completedMonthLesson := &domain.Lesson{
		ID:        uuid.New(),
		TeacherID: teacherID,
		ClientID:  client2ID,
		StartTime: startOfMonth.Add(24 * time.Hour),
		EndTime:   startOfMonth.Add(26 * time.Hour), // 2h * 1500 = 3000 факт (0% comm)
		Format:    domain.FormatIndividual,
		Status:    domain.StatusCompleted,
	}

	futureMonthLesson := &domain.Lesson{
		ID:        uuid.New(),
		TeacherID: teacherID,
		ClientID:  client1ID,
		StartTime: now.Add(2 * time.Hour),
		EndTime:   now.Add(3 * time.Hour), // 1h * 2000 = 2000 - 20% = 1600 прогноз
		Format:    domain.FormatIndividual,
		Status:    domain.StatusScheduled,
	}

	clientRepo.On("ListByTeacherID", ctx, teacherID).Return([]*domain.Client{client1, client2}, nil)
	classroomRepo.On("List", ctx).Return([]*domain.Classroom{classroom}, nil)

	lessonRepo.On("List", ctx, mock.MatchedBy(func(f schedule.LessonFilter) bool {
		return f.From != nil && f.From.Equal(startOfToday)
	})).Return([]*domain.Lesson{todayLesson}, nil)

	lessonRepo.On("List", ctx, mock.MatchedBy(func(f schedule.LessonFilter) bool {
		return f.From != nil && f.From.Equal(startOfMonth)
	})).Return([]*domain.Lesson{completedMonthLesson, futureMonthLesson}, nil)

	lessonRepo.On("List", ctx, mock.MatchedBy(func(f schedule.LessonFilter) bool {
		// недельный фильтр
		return f.From != nil && !f.From.Equal(startOfToday) && !f.From.Equal(startOfMonth)
	})).Return([]*domain.Lesson{todayLesson, futureMonthLesson}, nil)

	summary, err := svc.GetSummary(ctx, teacherID)
	require.NoError(t, err)

	require.Len(t, summary.TodayLessons, 1)
	assert.Equal(t, todayLesson.ID, summary.TodayLessons[0].ID)
	assert.Equal(t, "Алексей Иванов", summary.TodayLessons[0].ClientName)
	assert.Equal(t, "offline", *summary.TodayLessons[0].LocationType)
	assert.Equal(t, "Аудитория 101", *summary.TodayLessons[0].ClassroomName)
	assert.Equal(t, "#3B82F6", *summary.TodayLessons[0].ClassroomColor)

	assert.Equal(t, 3000.0, summary.FinancialSnapshot.MonthEarned)
	assert.Equal(t, 1600.0, summary.FinancialSnapshot.MonthForecast)
	assert.Equal(t, 4000.0, summary.FinancialSnapshot.TotalDebts)
	assert.Equal(t, 2, summary.FinancialSnapshot.ActiveClientsCount)
	assert.Equal(t, 2.0, summary.FinancialSnapshot.WeeklyHours) // 1h + 1h
}
