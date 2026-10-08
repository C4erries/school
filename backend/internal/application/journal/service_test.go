package journal_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/application/journal"
	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
)

type mockJournalRepo struct {
	mock.Mock
}

func (m *mockJournalRepo) GetByLessonID(ctx context.Context, lessonID uuid.UUID) (*domain.LessonJournal, error) {
	args := m.Called(ctx, lessonID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LessonJournal), args.Error(1)
}

func (m *mockJournalRepo) Upsert(ctx context.Context, j *domain.LessonJournal) error {
	args := m.Called(ctx, j)
	return args.Error(0)
}

func (m *mockJournalRepo) ListByClientID(ctx context.Context, clientID, teacherID uuid.UUID) ([]*domain.LessonJournal, error) {
	args := m.Called(ctx, clientID, teacherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LessonJournal), args.Error(1)
}

type mockHomeworkRepo struct {
	mock.Mock
}

func (m *mockHomeworkRepo) Create(ctx context.Context, hw *domain.HomeworkAssignment) error {
	args := m.Called(ctx, hw)
	return args.Error(0)
}

func (m *mockHomeworkRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.HomeworkAssignment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.HomeworkAssignment), args.Error(1)
}

func (m *mockHomeworkRepo) Update(ctx context.Context, hw *domain.HomeworkAssignment) error {
	args := m.Called(ctx, hw)
	return args.Error(0)
}

func (m *mockHomeworkRepo) Delete(ctx context.Context, id, teacherID uuid.UUID) error {
	args := m.Called(ctx, id, teacherID)
	return args.Error(0)
}

func (m *mockHomeworkRepo) ListByClientID(ctx context.Context, clientID, teacherID uuid.UUID, statusFilter *domain.HomeworkStatus) ([]*domain.HomeworkAssignment, error) {
	args := m.Called(ctx, clientID, teacherID, statusFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.HomeworkAssignment), args.Error(1)
}

func (m *mockHomeworkRepo) ListByAssignedLessonID(ctx context.Context, lessonID uuid.UUID) ([]*domain.HomeworkAssignment, error) {
	args := m.Called(ctx, lessonID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.HomeworkAssignment), args.Error(1)
}

func (m *mockHomeworkRepo) ListDueByLessonOrDate(ctx context.Context, clientID uuid.UUID, targetDate time.Time) ([]*domain.HomeworkAssignment, error) {
	args := m.Called(ctx, clientID, targetDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.HomeworkAssignment), args.Error(1)
}

type mockScheduleProvider struct {
	mock.Mock
}

func (m *mockScheduleProvider) GetLesson(ctx context.Context, lessonID uuid.UUID) (*domain.Lesson, error) {
	args := m.Called(ctx, lessonID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Lesson), args.Error(1)
}

func (m *mockScheduleProvider) FindVirtualLesson(ctx context.Context, lessonID uuid.UUID, teacherID *uuid.UUID) (*domain.LessonSeries, *schedule.OccurrenceSlot, error) {
	args := m.Called(ctx, lessonID, teacherID)
	var series *domain.LessonSeries
	var slot *schedule.OccurrenceSlot
	if args.Get(0) != nil {
		series = args.Get(0).(*domain.LessonSeries)
	}
	if args.Get(1) != nil {
		slot = args.Get(1).(*schedule.OccurrenceSlot)
	}
	return series, slot, args.Error(2)
}

func (m *mockScheduleProvider) EnsurePhysicalLesson(ctx context.Context, lessonID uuid.UUID, teacherID uuid.UUID) (*domain.Lesson, error) {
	args := m.Called(ctx, lessonID, teacherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Lesson), args.Error(1)
}

type mockClientProvider struct {
	mock.Mock
}

func (m *mockClientProvider) GetClient(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Client), args.Error(1)
}

func TestJournalService_GetLessonJournalBundle(t *testing.T) {
	ctx := context.Background()
	teacherID := uuid.New()
	clientID := uuid.New()
	lessonID := uuid.New()
	now := time.Now().UTC()

	t.Run("Physical lesson exists with journal and homeworks", func(t *testing.T) {
		jRepo := new(mockJournalRepo)
		hwRepo := new(mockHomeworkRepo)
		schedProv := new(mockScheduleProvider)
		clientProv := new(mockClientProvider)
		svc := journal.NewService(jRepo, hwRepo, schedProv, clientProv)

		physLesson := &domain.Lesson{
			ID:        lessonID,
			TeacherID: teacherID,
			ClientID:  clientID,
			StartTime: now,
		}
		score := 5
		existingJournal := &domain.LessonJournal{
			ID:               uuid.New(),
			LessonID:         lessonID,
			ClientID:         clientID,
			TeacherID:        teacherID,
			Topic:            "Тригонометрия",
			PerformanceScore: &score,
		}
		hws := []*domain.HomeworkAssignment{
			{ID: uuid.New(), Title: "Задание 1"},
		}

		schedProv.On("GetLesson", ctx, lessonID).Return(physLesson, nil).Once()
		jRepo.On("GetByLessonID", ctx, lessonID).Return(existingJournal, nil).Once()
		hwRepo.On("ListByAssignedLessonID", ctx, lessonID).Return(hws, nil).Once()
		hwRepo.On("ListDueByLessonOrDate", ctx, clientID, now).Return(hws, nil).Once()

		bundle, err := svc.GetLessonJournalBundle(ctx, lessonID, teacherID)
		require.NoError(t, err)
		assert.Equal(t, existingJournal, bundle.Journal)
		assert.Len(t, bundle.AssignedHomeworks, 1)
		assert.Len(t, bundle.DueHomeworks, 1)
	})

	t.Run("Virtual lesson returns empty bundle without journal", func(t *testing.T) {
		jRepo := new(mockJournalRepo)
		hwRepo := new(mockHomeworkRepo)
		schedProv := new(mockScheduleProvider)
		clientProv := new(mockClientProvider)
		svc := journal.NewService(jRepo, hwRepo, schedProv, clientProv)

		schedProv.On("GetLesson", ctx, lessonID).Return(nil, domain.ErrLessonNotFound).Once()
		series := &domain.LessonSeries{
			ID:        uuid.New(),
			TeacherID: teacherID,
			ClientID:  clientID,
		}
		slot := &schedule.OccurrenceSlot{
			StartTime: now,
		}
		schedProv.On("FindVirtualLesson", ctx, lessonID, &teacherID).Return(series, slot, nil).Once()
		hwRepo.On("ListDueByLessonOrDate", ctx, clientID, now).Return([]*domain.HomeworkAssignment{}, nil).Once()

		bundle, err := svc.GetLessonJournalBundle(ctx, lessonID, teacherID)
		require.NoError(t, err)
		assert.Nil(t, bundle.Journal)
		assert.Empty(t, bundle.AssignedHomeworks)
		assert.Empty(t, bundle.DueHomeworks)
	})
}

func TestJournalService_UpsertLessonJournal(t *testing.T) {
	ctx := context.Background()
	teacherID := uuid.New()
	clientID := uuid.New()
	lessonID := uuid.New()

	t.Run("Success materializes virtual and saves journal", func(t *testing.T) {
		jRepo := new(mockJournalRepo)
		hwRepo := new(mockHomeworkRepo)
		schedProv := new(mockScheduleProvider)
		clientProv := new(mockClientProvider)
		svc := journal.NewService(jRepo, hwRepo, schedProv, clientProv)

		physLesson := &domain.Lesson{
			ID:        lessonID,
			TeacherID: teacherID,
			ClientID:  clientID,
		}
		schedProv.On("EnsurePhysicalLesson", ctx, lessonID, teacherID).Return(physLesson, nil).Once()
		jRepo.On("Upsert", ctx, mock.AnythingOfType("*domain.LessonJournal")).Return(nil).Once()

		score := 4
		res, err := svc.UpsertLessonJournal(ctx, teacherID, lessonID, journal.UpsertJournalInput{
			Topic:            "Производные",
			Notes:            "Разобрали формулу",
			PerformanceScore: &score,
		})
		require.NoError(t, err)
		assert.Equal(t, "Производные", res.Topic)
		assert.Equal(t, 4, *res.PerformanceScore)
	})
}
