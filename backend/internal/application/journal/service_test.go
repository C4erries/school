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

func (m *mockScheduleProvider) GetUpcomingLesson(ctx context.Context, clientID, teacherID uuid.UUID) (*domain.Lesson, error) {
	args := m.Called(ctx, clientID, teacherID)
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

type mockNotesRepo struct {
	mock.Mock
}

func (m *mockNotesRepo) Create(ctx context.Context, note *domain.ClientNote) error {
	args := m.Called(ctx, note)
	return args.Error(0)
}

func (m *mockNotesRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ClientNote, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ClientNote), args.Error(1)
}

func (m *mockNotesRepo) Delete(ctx context.Context, id, teacherID uuid.UUID) error {
	args := m.Called(ctx, id, teacherID)
	return args.Error(0)
}

func (m *mockNotesRepo) ListByClientID(ctx context.Context, clientID, teacherID uuid.UUID) ([]*domain.ClientNote, error) {
	args := m.Called(ctx, clientID, teacherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ClientNote), args.Error(1)
}

func TestJournalService_ClientNotesAndStream(t *testing.T) {
	ctx := context.Background()
	teacherID := uuid.New()
	clientID := uuid.New()
	otherTeacherID := uuid.New()

	t.Run("CreateClientNote success", func(t *testing.T) {
		jRepo := new(mockJournalRepo)
		hwRepo := new(mockHomeworkRepo)
		schedProv := new(mockScheduleProvider)
		clientProv := new(mockClientProvider)
		notesRepo := new(mockNotesRepo)
		svc := journal.NewService(jRepo, hwRepo, schedProv, clientProv, notesRepo)

		clientProv.On("GetClient", ctx, clientID).Return(&domain.Client{
			ID:        clientID,
			TeacherID: teacherID,
		}, nil).Once()
		notesRepo.On("Create", ctx, mock.AnythingOfType("*domain.ClientNote")).Return(nil).Once()

		note, err := svc.CreateClientNote(ctx, teacherID, clientID, "Тестовая заметка")
		require.NoError(t, err)
		assert.Equal(t, "Тестовая заметка", note.Content)
		assert.Equal(t, clientID, note.ClientID)
		assert.Equal(t, teacherID, note.TeacherID)
	})

	t.Run("CreateClientNote forbidden for another teacher", func(t *testing.T) {
		jRepo := new(mockJournalRepo)
		hwRepo := new(mockHomeworkRepo)
		schedProv := new(mockScheduleProvider)
		clientProv := new(mockClientProvider)
		notesRepo := new(mockNotesRepo)
		svc := journal.NewService(jRepo, hwRepo, schedProv, clientProv, notesRepo)

		clientProv.On("GetClient", ctx, clientID).Return(&domain.Client{
			ID:        clientID,
			TeacherID: otherTeacherID,
		}, nil).Once()

		_, err := svc.CreateClientNote(ctx, teacherID, clientID, "Заметка")
		require.ErrorIs(t, err, domain.ErrForbidden)
	})

	t.Run("DeleteClientNote success", func(t *testing.T) {
		jRepo := new(mockJournalRepo)
		hwRepo := new(mockHomeworkRepo)
		schedProv := new(mockScheduleProvider)
		clientProv := new(mockClientProvider)
		notesRepo := new(mockNotesRepo)
		svc := journal.NewService(jRepo, hwRepo, schedProv, clientProv, notesRepo)

		noteID := uuid.New()
		notesRepo.On("Delete", ctx, noteID, teacherID).Return(nil).Once()

		err := svc.DeleteClientNote(ctx, noteID, teacherID)
		require.NoError(t, err)
	})

	t.Run("GetClientStudyStream returns sorted items and upcoming lesson", func(t *testing.T) {
		jRepo := new(mockJournalRepo)
		hwRepo := new(mockHomeworkRepo)
		schedProv := new(mockScheduleProvider)
		clientProv := new(mockClientProvider)
		notesRepo := new(mockNotesRepo)
		svc := journal.NewService(jRepo, hwRepo, schedProv, clientProv, notesRepo)

		clientProv.On("GetClient", ctx, clientID).Return(&domain.Client{
			ID:        clientID,
			TeacherID: teacherID,
		}, nil).Once()

		t1 := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
		t2 := time.Date(2026, 10, 2, 10, 0, 0, 0, time.UTC)
		t3 := time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)

		lessonID := uuid.New()
		journalEntry := &domain.LessonJournal{
			ID:        uuid.New(),
			LessonID:  lessonID,
			ClientID:  clientID,
			TeacherID: teacherID,
			Topic:     "Теорема Пифагора",
			CreatedAt: t2,
		}

		noteEntry := &domain.ClientNote{
			ID:        uuid.New(),
			ClientID:  clientID,
			TeacherID: teacherID,
			Content:   "Заметки до урока",
			CreatedAt: t1,
		}

		upcoming := &domain.Lesson{
			ID:        uuid.New(),
			TeacherID: teacherID,
			ClientID:  clientID,
			Title:     "Следующий урок",
			StartTime: t3,
			Status:    domain.StatusScheduled,
		}

		jRepo.On("ListByClientID", ctx, clientID, teacherID).Return([]*domain.LessonJournal{journalEntry}, nil).Once()
		hwRepo.On("ListByAssignedLessonID", ctx, lessonID).Return([]*domain.HomeworkAssignment{}, nil).Once()
		hwRepo.On("ListDueByLessonOrDate", ctx, clientID, t2).Return([]*domain.HomeworkAssignment{}, nil).Once()
		notesRepo.On("ListByClientID", ctx, clientID, teacherID).Return([]*domain.ClientNote{noteEntry}, nil).Once()
		schedProv.On("GetUpcomingLesson", ctx, clientID, teacherID).Return(upcoming, nil).Once()

		stream, err := svc.GetClientStudyStream(ctx, teacherID, clientID)
		require.NoError(t, err)
		assert.Equal(t, clientID, stream.ClientID)
		assert.Equal(t, upcoming, stream.UpcomingLesson)
		require.Len(t, stream.Items, 2)
		// Проверяем сортировку ASC (t1 раньше t2)
		assert.Equal(t, "note", stream.Items[0].Type)
		assert.Equal(t, t1, stream.Items[0].Timestamp)
		assert.Equal(t, "lesson_report", stream.Items[1].Type)
		assert.Equal(t, t2, stream.Items[1].Timestamp)
	})
}
