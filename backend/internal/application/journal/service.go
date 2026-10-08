package journal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
)

// JournalRepository определяет контракт хранилища отчетов по урокам.
type JournalRepository interface {
	GetByLessonID(ctx context.Context, lessonID uuid.UUID) (*domain.LessonJournal, error)
	Upsert(ctx context.Context, journal *domain.LessonJournal) error
	ListByClientID(ctx context.Context, clientID, teacherID uuid.UUID) ([]*domain.LessonJournal, error)
}

// HomeworkRepository определяет контракт хранилища домашних заданий.
type HomeworkRepository interface {
	Create(ctx context.Context, hw *domain.HomeworkAssignment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.HomeworkAssignment, error)
	Update(ctx context.Context, hw *domain.HomeworkAssignment) error
	Delete(ctx context.Context, id, teacherID uuid.UUID) error
	ListByClientID(ctx context.Context, clientID, teacherID uuid.UUID, statusFilter *domain.HomeworkStatus) ([]*domain.HomeworkAssignment, error)
	ListByAssignedLessonID(ctx context.Context, lessonID uuid.UUID) ([]*domain.HomeworkAssignment, error)
	ListDueByLessonOrDate(ctx context.Context, clientID uuid.UUID, targetDate time.Time) ([]*domain.HomeworkAssignment, error)
}

// ScheduleProvider контракт взаимодействия с сервисом расписания для уроков и серий.
type ScheduleProvider interface {
	GetLesson(ctx context.Context, lessonID uuid.UUID) (*domain.Lesson, error)
	FindVirtualLesson(ctx context.Context, lessonID uuid.UUID, teacherID *uuid.UUID) (*domain.LessonSeries, *schedule.OccurrenceSlot, error)
	EnsurePhysicalLesson(ctx context.Context, lessonID uuid.UUID, teacherID uuid.UUID) (*domain.Lesson, error)
}

// ClientProvider контракт доступа к клиентам для проверки прав доступа.
type ClientProvider interface {
	GetClient(ctx context.Context, id uuid.UUID) (*domain.Client, error)
}

// LessonJournalBundle агрегирует журнал урока и связанные домашние задания.
type LessonJournalBundle struct {
	Journal            *domain.LessonJournal
	AssignedHomeworks  []*domain.HomeworkAssignment
	DueHomeworks       []*domain.HomeworkAssignment
}

// UpsertJournalInput параметры создания или обновления отчета по уроку.
type UpsertJournalInput struct {
	Topic            string
	Notes            string
	PerformanceScore *int
}

// CreateHomeworkInput входные данные для выдачи домашнего задания.
type CreateHomeworkInput struct {
	ClientID         uuid.UUID
	TeacherID        uuid.UUID
	AssignedLessonID *uuid.UUID
	Title            string
	Description      string
	DueDate          *time.Time
}

// UpdateHomeworkStatusInput входные данные для изменения статуса и рецензии ДЗ.
type UpdateHomeworkStatusInput struct {
	ID          uuid.UUID
	TeacherID   uuid.UUID
	Status      domain.HomeworkStatus
	ReviewNotes string
}

// Service реализует бизнес-логику дневника занятий и домашних заданий.
type Service struct {
	journalRepo      JournalRepository
	homeworkRepo     HomeworkRepository
	scheduleProvider ScheduleProvider
	clientProvider   ClientProvider
}

// NewService создает новый экземпляр сервиса дневника и домашних заданий.
func NewService(
	journalRepo JournalRepository,
	homeworkRepo HomeworkRepository,
	scheduleProvider ScheduleProvider,
	clientProvider ClientProvider,
) *Service {
	return &Service{
		journalRepo:      journalRepo,
		homeworkRepo:     homeworkRepo,
		scheduleProvider: scheduleProvider,
		clientProvider:   clientProvider,
	}
}

// GetLessonJournalBundle возвращает отчет по уроку, выданные ДЗ и ДЗ к проверке на этом уроке.
func (s *Service) GetLessonJournalBundle(ctx context.Context, lessonID, teacherID uuid.UUID) (*LessonJournalBundle, error) {
	// 1. Проверяем физический урок
	lesson, err := s.scheduleProvider.GetLesson(ctx, lessonID)
	if err == nil {
		if lesson.TeacherID != teacherID {
			return nil, domain.ErrUnauthorizedLessonAction
		}

		journal, jErr := s.journalRepo.GetByLessonID(ctx, lessonID)
		if jErr != nil && !errors.Is(jErr, domain.ErrJournalNotFound) {
			return nil, fmt.Errorf("get journal by lesson: %w", jErr)
		}

		assignedHws, aErr := s.homeworkRepo.ListByAssignedLessonID(ctx, lessonID)
		if aErr != nil {
			return nil, fmt.Errorf("list assigned homeworks: %w", aErr)
		}

		dueHws, dErr := s.homeworkRepo.ListDueByLessonOrDate(ctx, lesson.ClientID, lesson.StartTime)
		if dErr != nil {
			return nil, fmt.Errorf("list due homeworks: %w", dErr)
		}

		return &LessonJournalBundle{
			Journal:           journal,
			AssignedHomeworks: assignedHws,
			DueHomeworks:      dueHws,
		}, nil
	}

	if !errors.Is(err, domain.ErrLessonNotFound) {
		return nil, fmt.Errorf("get lesson by id: %w", err)
	}

	// 2. Если физического урока нет, проверяем виртуальный слот серии
	series, slot, vErr := s.scheduleProvider.FindVirtualLesson(ctx, lessonID, &teacherID)
	if vErr == nil && series != nil && slot != nil {
		if series.TeacherID != teacherID {
			return nil, domain.ErrUnauthorizedLessonAction
		}

		dueHws, dErr := s.homeworkRepo.ListDueByLessonOrDate(ctx, series.ClientID, slot.StartTime)
		if dErr != nil {
			return nil, fmt.Errorf("list due homeworks for virtual slot: %w", dErr)
		}

		return &LessonJournalBundle{
			Journal:           nil,
			AssignedHomeworks: make([]*domain.HomeworkAssignment, 0),
			DueHomeworks:      dueHws,
		}, nil
	}

	return nil, domain.ErrLessonNotFound
}

// UpsertLessonJournal сохраняет отчет по уроку. Если урок виртуальный, он автоматически материализуется.
func (s *Service) UpsertLessonJournal(
	ctx context.Context,
	teacherID, lessonID uuid.UUID,
	input UpsertJournalInput,
) (*domain.LessonJournal, error) {
	lesson, err := s.scheduleProvider.EnsurePhysicalLesson(ctx, lessonID, teacherID)
	if err != nil {
		return nil, err
	}

	journal, err := domain.NewLessonJournal(
		lesson.ID,
		lesson.ClientID,
		lesson.TeacherID,
		input.Topic,
		input.Notes,
		input.PerformanceScore,
	)
	if err != nil {
		return nil, err
	}

	if err := s.journalRepo.Upsert(ctx, journal); err != nil {
		return nil, fmt.Errorf("upsert lesson journal: %w", err)
	}

	return journal, nil
}

// GetClientJournals возвращает хронологический таймлайн отчетов по занятиям ученика.
func (s *Service) GetClientJournals(ctx context.Context, clientID, teacherID uuid.UUID) ([]*domain.LessonJournal, error) {
	client, err := s.clientProvider.GetClient(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if client.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}

	return s.journalRepo.ListByClientID(ctx, clientID, teacherID)
}

// ListClientHomework возвращает список всех домашних заданий ученика с фильтром по статусу.
func (s *Service) ListClientHomework(
	ctx context.Context,
	clientID, teacherID uuid.UUID,
	statusFilter *domain.HomeworkStatus,
) ([]*domain.HomeworkAssignment, error) {
	client, err := s.clientProvider.GetClient(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if client.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}

	return s.homeworkRepo.ListByClientID(ctx, clientID, teacherID, statusFilter)
}

// CreateHomework создает и выдает домашнее задание ученику.
func (s *Service) CreateHomework(ctx context.Context, input CreateHomeworkInput) (*domain.HomeworkAssignment, error) {
	client, err := s.clientProvider.GetClient(ctx, input.ClientID)
	if err != nil {
		return nil, err
	}
	if client.TeacherID != input.TeacherID {
		return nil, domain.ErrForbidden
	}

	var assignedLessonID *uuid.UUID
	if input.AssignedLessonID != nil {
		lesson, err := s.scheduleProvider.EnsurePhysicalLesson(ctx, *input.AssignedLessonID, input.TeacherID)
		if err != nil {
			return nil, err
		}
		assignedLessonID = &lesson.ID
	}

	hw, err := domain.NewHomeworkAssignment(
		input.ClientID,
		input.TeacherID,
		assignedLessonID,
		input.Title,
		input.Description,
		input.DueDate,
	)
	if err != nil {
		return nil, err
	}

	if err := s.homeworkRepo.Create(ctx, hw); err != nil {
		return nil, fmt.Errorf("create homework assignment: %w", err)
	}

	return hw, nil
}

// UpdateHomeworkStatus изменяет статус выполнения задания и рецензию преподавателя.
func (s *Service) UpdateHomeworkStatus(
	ctx context.Context,
	input UpdateHomeworkStatusInput,
) (*domain.HomeworkAssignment, error) {
	hw, err := s.homeworkRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if hw.TeacherID != input.TeacherID {
		return nil, domain.ErrForbidden
	}

	if err := hw.UpdateStatus(input.Status, input.ReviewNotes); err != nil {
		return nil, err
	}

	if err := s.homeworkRepo.Update(ctx, hw); err != nil {
		return nil, fmt.Errorf("update homework assignment: %w", err)
	}

	return hw, nil
}

// DeleteHomework удаляет домашнее задание преподавателя.
func (s *Service) DeleteHomework(ctx context.Context, id, teacherID uuid.UUID) error {
	return s.homeworkRepo.Delete(ctx, id, teacherID)
}
