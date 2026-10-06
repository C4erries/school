package schedule_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
)

type MockSeriesRepo struct {
	mock.Mock
}

func (m *MockSeriesRepo) Create(ctx context.Context, s *domain.LessonSeries) error {
	return m.Called(ctx, s).Error(0)
}
func (m *MockSeriesRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.LessonSeries, error) {
	args := m.Called(ctx, id)
	if s := args.Get(0); s != nil {
		return s.(*domain.LessonSeries), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockSeriesRepo) Update(ctx context.Context, s *domain.LessonSeries) error {
	return m.Called(ctx, s).Error(0)
}
func (m *MockSeriesRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockSeriesRepo) ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.LessonSeries, error) {
	args := m.Called(ctx, teacherID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.LessonSeries), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestScheduleService_RecurringLessons(t *testing.T) {
	teacherID := uuid.New()
	clientID := uuid.New()
	seriesID := uuid.New()

	lessonRepo := new(MockLessonRepository)
	classroomRepo := new(MockClassroomRepository)
	clientRepo := new(MockClientRepository)
	subRepo := new(MockSubscriptionRepository)
	seriesRepo := new(MockSeriesRepo)

	svc := schedule.NewService(classroomRepo, lessonRepo, clientRepo, subRepo, seriesRepo)

	client := &domain.Client{
		ID:        clientID,
		TeacherID: teacherID,
		Name:      "Ученик Серии",
	}

	t.Run("CreateSeries_success", func(t *testing.T) {
		clientRepo.On("GetByID", mock.Anything, clientID).Return(client, nil).Once()
		seriesRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.LessonSeries) bool {
			return s.TeacherID == teacherID && s.ClientID == clientID && s.RRULE == "FREQ=WEEKLY;BYDAY=TU,TH"
		})).Return(nil).Once()

		series, err := svc.CreateSeries(context.Background(), schedule.CreateSeriesInput{
			TeacherID:       teacherID,
			ClientID:        clientID,
			Title:           "Регулярная физика",
			RRULE:           "FREQ=WEEKLY;BYDAY=TU,TH",
			StartTimeOfDay:  "17:00",
			DurationMinutes: 60,
			Format:          domain.FormatIndividual,
			StartDate:       time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		})
		require.NoError(t, err)
		assert.Equal(t, "Регулярная физика", series.Title)
		assert.Equal(t, "17:00", series.StartTimeOfDay)
	})

	t.Run("ListLessonsWithRecurring_generates_virtual_slots_and_merges", func(t *testing.T) {
		from := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC) // Понедельник
		to := time.Date(2026, 10, 11, 23, 59, 59, 0, time.UTC) // Воскресенье

		// В неделе 5-11 октября 2026:
		// Вт 6 окт 17:00
		// Чт 8 окт 17:00
		series := &domain.LessonSeries{
			ID:              seriesID,
			TeacherID:       teacherID,
			ClientID:        clientID,
			Title:           "Физика",
			RRULE:           "FREQ=WEEKLY;BYDAY=TU,TH",
			StartTimeOfDay:  "17:00",
			DurationMinutes: 60,
			Format:          domain.FormatIndividual,
			StartDate:       time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		}

		lessonRepo.On("List", mock.Anything, mock.Anything).Return([]*domain.Lesson{}, nil).Once()
		seriesRepo.On("ListByTeacherID", mock.Anything, teacherID).Return([]*domain.LessonSeries{series}, nil).Once()

		lessons, err := svc.ListLessons(context.Background(), schedule.LessonFilter{
			TeacherID: &teacherID,
			From:      &from,
			To:        &to,
		})
		require.NoError(t, err)
		require.Len(t, lessons, 2)
		assert.Equal(t, time.Tuesday, lessons[0].StartTime.Weekday())
		assert.Equal(t, time.Thursday, lessons[1].StartTime.Weekday())
		assert.Equal(t, &seriesID, lessons[0].SeriesID)
	})

	t.Run("CompleteRecurringLesson_virtual_materializes_and_deducts_subscription", func(t *testing.T) {
		series := &domain.LessonSeries{
			ID:              seriesID,
			TeacherID:       teacherID,
			ClientID:        clientID,
			Title:           "Физика",
			RRULE:           "FREQ=WEEKLY;BYDAY=TU,TH",
			StartTimeOfDay:  "17:00",
			DurationMinutes: 60,
			Format:          domain.FormatIndividual,
			StartDate:       time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		}

		occDate := time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)
		lessonRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, domain.ErrLessonNotFound).Once()
		seriesRepo.On("GetByID", mock.Anything, seriesID).Return(series, nil).Once()
		lessonRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *domain.Lesson) bool {
			return l.Status == domain.StatusCompleted && l.SeriesID != nil && *l.SeriesID == seriesID
		})).Return(nil).Once()
		subRepo.On("GetByClientID", mock.Anything, clientID).Return([]*domain.ClientSubscription{}, nil).Once()
		subRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.ClientSubscription) bool {
			return s.Balance == -1.0
		})).Return(nil).Once()

		mat, err := svc.CompleteRecurringLesson(context.Background(), uuid.New(), &seriesID, &occDate, teacherID, domain.RoleTeacher)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusCompleted, mat.Status)
		assert.Equal(t, &seriesID, mat.SeriesID)
	})

	t.Run("CancelRecurringLesson_ScopeThisOnly_materializes_cancelled_lesson", func(t *testing.T) {
		series := &domain.LessonSeries{
			ID:              seriesID,
			TeacherID:       teacherID,
			ClientID:        clientID,
			Title:           "Физика",
			RRULE:           "FREQ=WEEKLY;BYDAY=TU,TH",
			StartTimeOfDay:  "17:00",
			DurationMinutes: 60,
			Format:          domain.FormatIndividual,
			StartDate:       time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		}

		occDate := time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
		lessonRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, domain.ErrLessonNotFound).Once()
		seriesRepo.On("GetByID", mock.Anything, seriesID).Return(series, nil).Once()
		lessonRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *domain.Lesson) bool {
			return l.Status == domain.StatusCancelled && l.CancelReason == "Отмена учеником"
		})).Return(nil).Once()

		err := svc.CancelRecurringLesson(context.Background(), uuid.New(), &seriesID, &occDate, domain.ScopeThisOnly, "Отмена учеником", teacherID, domain.RoleTeacher)
		require.NoError(t, err)
	})
}
