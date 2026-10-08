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

// NotesRepository определяет контракт хранилища свободных заметок.
type NotesRepository interface {
	Create(ctx context.Context, note *domain.ClientNote) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ClientNote, error)
	Delete(ctx context.Context, id, teacherID uuid.UUID) error
	ListByClientID(ctx context.Context, clientID, teacherID uuid.UUID) ([]*domain.ClientNote, error)
}

// ScheduleProvider контракт взаимодействия с сервисом расписания для уроков и серий.
type ScheduleProvider interface {
	GetLesson(ctx context.Context, lessonID uuid.UUID) (*domain.Lesson, error)
	FindVirtualLesson(ctx context.Context, lessonID uuid.UUID, teacherID *uuid.UUID) (*domain.LessonSeries, *schedule.OccurrenceSlot, error)
	EnsurePhysicalLesson(ctx context.Context, lessonID uuid.UUID, teacherID uuid.UUID) (*domain.Lesson, error)
	GetUpcomingLesson(ctx context.Context, clientID, teacherID uuid.UUID) (*domain.Lesson, error)
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

// StudyStreamItem объединяет отчет по уроку или свободную заметку для ленты.
type StudyStreamItem struct {
	ID           uuid.UUID
	Type         string // "lesson_report" или "note"
	Timestamp    time.Time
	LessonReport *LessonJournalBundle
	Note         *domain.ClientNote
}

// StudyStreamResult содержит ленту занятий/заметок ученика и ближайший запланированный урок.
type StudyStreamResult struct {
	ClientID       uuid.UUID
	Items          []*StudyStreamItem
	UpcomingLesson *domain.Lesson
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
	notesRepo        NotesRepository
	scheduleProvider ScheduleProvider
	clientProvider   ClientProvider
}

// NewService создает новый экземпляр сервиса дневника и домашних заданий.
func NewService(
	journalRepo JournalRepository,
	homeworkRepo HomeworkRepository,
	scheduleProvider ScheduleProvider,
	clientProvider ClientProvider,
	notesRepo ...NotesRepository,
) *Service {
	var nRepo NotesRepository
	if len(notesRepo) > 0 {
		nRepo = notesRepo[0]
	}
	return &Service{
		journalRepo:      journalRepo,
		homeworkRepo:     homeworkRepo,
		notesRepo:        nRepo,
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

// CreateClientNote сохраняет новую свободную заметку по ученику.
func (s *Service) CreateClientNote(ctx context.Context, teacherID, clientID uuid.UUID, content string) (*domain.ClientNote, error) {
	if s.notesRepo == nil {
		return nil, errors.New("notes repository not configured")
	}

	client, err := s.clientProvider.GetClient(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if client.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}

	note, err := domain.NewClientNote(clientID, teacherID, content)
	if err != nil {
		return nil, err
	}

	if err := s.notesRepo.Create(ctx, note); err != nil {
		return nil, fmt.Errorf("create client note: %w", err)
	}

	return note, nil
}

// DeleteClientNote удаляет свободную заметку по ученику.
func (s *Service) DeleteClientNote(ctx context.Context, id, teacherID uuid.UUID) error {
	if s.notesRepo == nil {
		return errors.New("notes repository not configured")
	}

	return s.notesRepo.Delete(ctx, id, teacherID)
}

// GetClientStudyStream возвращает объединенную хронологическую ленту отчетов по урокам и заметок, а также ближайший запланированный урок.
func (s *Service) GetClientStudyStream(ctx context.Context, teacherID, clientID uuid.UUID) (*StudyStreamResult, error) {
	client, err := s.clientProvider.GetClient(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if client.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}

	// 1. Получаем все отчеты по урокам данного клиента
	journals, err := s.journalRepo.ListByClientID(ctx, clientID, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list journals by client: %w", err)
	}

	items := make([]*StudyStreamItem, 0, len(journals))
	for _, j := range journals {
		assignedHws, aErr := s.homeworkRepo.ListByAssignedLessonID(ctx, j.LessonID)
		if aErr != nil {
			assignedHws = []*domain.HomeworkAssignment{}
		}

		dueHws, dErr := s.homeworkRepo.ListDueByLessonOrDate(ctx, clientID, j.CreatedAt)
		if dErr != nil {
			dueHws = []*domain.HomeworkAssignment{}
		}

		bundle := &LessonJournalBundle{
			Journal:           j,
			AssignedHomeworks: assignedHws,
			DueHomeworks:      dueHws,
		}

		items = append(items, &StudyStreamItem{
			ID:           j.ID,
			Type:         "lesson_report",
			Timestamp:    j.CreatedAt,
			LessonReport: bundle,
		})
	}

	// 2. Получаем свободные заметки
	if s.notesRepo != nil {
		notes, nErr := s.notesRepo.ListByClientID(ctx, clientID, teacherID)
		if nErr != nil {
			return nil, fmt.Errorf("list notes by client: %w", nErr)
		}
		for _, note := range notes {
			items = append(items, &StudyStreamItem{
				ID:        note.ID,
				Type:      "note",
				Timestamp: note.CreatedAt,
				Note:      note,
			})
		}
	}

	// 3. Сортируем ленту строго Timestamp ASC (старые сверху, новые снизу)
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].Timestamp.After(items[j].Timestamp) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// 4. Получаем ближайший запланированный урок
	var upcoming *domain.Lesson
	if s.scheduleProvider != nil {
		up, uErr := s.scheduleProvider.GetUpcomingLesson(ctx, clientID, teacherID)
		if uErr == nil {
			upcoming = up
		}
	}

	return &StudyStreamResult{
		ClientID:       clientID,
		Items:          items,
		UpcomingLesson: upcoming,
	}, nil
}
