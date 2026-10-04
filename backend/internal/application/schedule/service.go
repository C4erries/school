package schedule

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
)

// LessonFilter параметры выборки списка уроков.
type LessonFilter struct {
	TeacherID   *uuid.UUID
	StudentID   *uuid.UUID
	ClassroomID *uuid.UUID
	Status      *domain.LessonStatus
	From        *time.Time
	To          *time.Time
}

// ClassroomRepository определяет контракт хранилища кабинетов.
type ClassroomRepository interface {
	Create(ctx context.Context, c *domain.Classroom) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Classroom, error)
	List(ctx context.Context) ([]*domain.Classroom, error)
	Update(ctx context.Context, c *domain.Classroom) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TeacherStudentRepository определяет контракт привязки учеников к преподавателям.
type TeacherStudentRepository interface {
	Create(ctx context.Context, ts *domain.TeacherStudent) error
	Delete(ctx context.Context, teacherID, studentID uuid.UUID) error
	IsAssigned(ctx context.Context, teacherID, studentID uuid.UUID) (bool, error)
	ListStudentsByTeacher(ctx context.Context, teacherID uuid.UUID) ([]*domain.User, error)
	ListTeachersByStudent(ctx context.Context, studentID uuid.UUID) ([]*domain.User, error)
}

// LessonRepository определяет контракт хранилища занятий.
type LessonRepository interface {
	Create(ctx context.Context, l *domain.Lesson) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Lesson, error)
	Update(ctx context.Context, l *domain.Lesson) error
	HasClassroomCollision(ctx context.Context, classroomID uuid.UUID, teacherID uuid.UUID, startTime, endTime time.Time, excludeLessonID *uuid.UUID) (bool, error)
	List(ctx context.Context, filter LessonFilter) ([]*domain.Lesson, error)
}

// UserRepository определяет контракт доступа к пользователям для валидации ролей.
type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

// DTO структуры входных данных для сервиса.
type CreateClassroomInput struct {
	Name        string
	Capacity    int
	Color       string
	Description string
}

type UpdateClassroomInput struct {
	ID          uuid.UUID
	Name        string
	Capacity    int
	Color       string
	Description string
}

type ScheduleLessonInput struct {
	TeacherID     uuid.UUID
	StudentID     uuid.UUID
	ClassroomID   *uuid.UUID
	StartTime     time.Time
	EndTime       time.Time
	Format        domain.LessonFormat
	LocationOrURL string
	Notes         string
}

// Service реализует бизнес-логику расписания, аудиторий и проведения уроков.
type Service struct {
	classroomRepo      ClassroomRepository
	teacherStudentRepo TeacherStudentRepository
	lessonRepo         LessonRepository
	userRepo           UserRepository
}

func NewService(
	classroomRepo ClassroomRepository,
	teacherStudentRepo TeacherStudentRepository,
	lessonRepo LessonRepository,
	userRepo UserRepository,
) *Service {
	return &Service{
		classroomRepo:      classroomRepo,
		teacherStudentRepo: teacherStudentRepo,
		lessonRepo:         lessonRepo,
		userRepo:           userRepo,
	}
}

// --- Кабинеты (Classrooms) ---

func (s *Service) CreateClassroom(ctx context.Context, input CreateClassroomInput) (*domain.Classroom, error) {
	name, err := domain.ValidateClassroomInput(input.Name, input.Capacity)
	if err != nil {
		return nil, err
	}

	color := strings.TrimSpace(input.Color)
	if color == "" {
		color = "#3B82F6"
	}

	classroom := &domain.Classroom{
		ID:          uuid.New(),
		Name:        name,
		Capacity:    input.Capacity,
		Color:       color,
		Description: strings.TrimSpace(input.Description),
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.classroomRepo.Create(ctx, classroom); err != nil {
		return nil, fmt.Errorf("create classroom: %w", err)
	}

	return classroom, nil
}

func (s *Service) GetClassroom(ctx context.Context, id uuid.UUID) (*domain.Classroom, error) {
	return s.classroomRepo.GetByID(ctx, id)
}

func (s *Service) ListClassrooms(ctx context.Context) ([]*domain.Classroom, error) {
	return s.classroomRepo.List(ctx)
}

func (s *Service) UpdateClassroom(ctx context.Context, input UpdateClassroomInput) (*domain.Classroom, error) {
	name, err := domain.ValidateClassroomInput(input.Name, input.Capacity)
	if err != nil {
		return nil, err
	}

	existing, err := s.classroomRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	color := strings.TrimSpace(input.Color)
	if color == "" {
		color = existing.Color
	}

	existing.Name = name
	existing.Capacity = input.Capacity
	existing.Color = color
	existing.Description = strings.TrimSpace(input.Description)

	if err := s.classroomRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("update classroom: %w", err)
	}

	return existing, nil
}

func (s *Service) DeleteClassroom(ctx context.Context, id uuid.UUID) error {
	return s.classroomRepo.Delete(ctx, id)
}

// --- Привязка учеников к преподавателям (Teacher-Student) ---

func (s *Service) AssignStudent(ctx context.Context, teacherID, studentID uuid.UUID) (*domain.TeacherStudent, error) {
	if teacherID == studentID {
		return nil, domain.ErrCannotAssignSelf
	}

	teacher, err := s.userRepo.GetByID(ctx, teacherID)
	if err != nil {
		return nil, err
	}
	if teacher.Role != domain.RoleTeacher && teacher.Role != domain.RoleOwner {
		return nil, domain.ErrInvalidTeacherRole
	}

	student, err := s.userRepo.GetByID(ctx, studentID)
	if err != nil {
		return nil, err
	}
	if student.Role != domain.RoleStudent {
		return nil, domain.ErrInvalidStudentRole
	}

	ts := &domain.TeacherStudent{
		ID:        uuid.New(),
		TeacherID: teacherID,
		StudentID: studentID,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.teacherStudentRepo.Create(ctx, ts); err != nil {
		return nil, err
	}

	return ts, nil
}

func (s *Service) UnassignStudent(ctx context.Context, teacherID, studentID uuid.UUID) error {
	return s.teacherStudentRepo.Delete(ctx, teacherID, studentID)
}

func (s *Service) ListTeacherStudents(ctx context.Context, teacherID uuid.UUID) ([]*domain.User, error) {
	return s.teacherStudentRepo.ListStudentsByTeacher(ctx, teacherID)
}

func (s *Service) ListStudentTeachers(ctx context.Context, studentID uuid.UUID) ([]*domain.User, error) {
	return s.teacherStudentRepo.ListTeachersByStudent(ctx, studentID)
}

// --- Уроки (Lessons) ---

func (s *Service) ScheduleLesson(ctx context.Context, input ScheduleLessonInput) (*domain.Lesson, error) {
	if !input.Format.IsValid() {
		return nil, domain.ErrInvalidLessonFormat
	}

	if err := domain.ValidateLessonTimes(input.StartTime, input.EndTime); err != nil {
		return nil, err
	}

	// Проверяем, закреплен ли ученик за преподавателем
	assigned, err := s.teacherStudentRepo.IsAssigned(ctx, input.TeacherID, input.StudentID)
	if err != nil {
		return nil, fmt.Errorf("check assignment: %w", err)
	}
	if !assigned {
		return nil, domain.ErrStudentNotAssignedToTeacher
	}

	var classroomID *uuid.UUID
	if input.Format == domain.FormatOffline {
		if input.ClassroomID == nil || *input.ClassroomID == uuid.Nil {
			return nil, domain.ErrClassroomRequiredForOffline
		}
		// Проверяем существование кабинета
		if _, err := s.classroomRepo.GetByID(ctx, *input.ClassroomID); err != nil {
			return nil, err
		}

		// Проверяем коллизию кабинета: два РАЗНЫХ преподавателя не могут занять один кабинет одновременно.
		collision, err := s.lessonRepo.HasClassroomCollision(
			ctx,
			*input.ClassroomID,
			input.TeacherID,
			input.StartTime,
			input.EndTime,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("check classroom collision: %w", err)
		}
		if collision {
			return nil, domain.ErrClassroomCollision
		}

		classroomID = input.ClassroomID
	}

	now := time.Now().UTC()
	lesson := &domain.Lesson{
		ID:            uuid.New(),
		TeacherID:     input.TeacherID,
		StudentID:     input.StudentID,
		ClassroomID:   classroomID,
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		Format:        input.Format,
		LocationOrURL: strings.TrimSpace(input.LocationOrURL),
		Status:        domain.StatusPendingConfirmation,
		Notes:         strings.TrimSpace(input.Notes),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.lessonRepo.Create(ctx, lesson); err != nil {
		return nil, fmt.Errorf("create lesson: %w", err)
	}

	return lesson, nil
}

func (s *Service) AcceptLesson(ctx context.Context, lessonID, callerID uuid.UUID) (*domain.Lesson, error) {
	lesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return nil, err
	}

	// Только ученик, которому назначен урок, может подтвердить его
	if lesson.StudentID != callerID {
		return nil, domain.ErrUnauthorizedLessonAction
	}

	if err := lesson.Accept(); err != nil {
		return nil, err
	}

	if err := s.lessonRepo.Update(ctx, lesson); err != nil {
		return nil, fmt.Errorf("update accepted lesson: %w", err)
	}

	return lesson, nil
}

func (s *Service) DeclineLesson(ctx context.Context, lessonID, callerID uuid.UUID, reason string) (*domain.Lesson, error) {
	lesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return nil, err
	}

	// Только ученик может отклонить предложенный урок
	if lesson.StudentID != callerID {
		return nil, domain.ErrUnauthorizedLessonAction
	}

	if err := lesson.Decline(reason); err != nil {
		return nil, err
	}

	if err := s.lessonRepo.Update(ctx, lesson); err != nil {
		return nil, fmt.Errorf("update declined lesson: %w", err)
	}

	return lesson, nil
}

func (s *Service) CompleteLesson(ctx context.Context, lessonID, callerID uuid.UUID, callerRole domain.Role) (*domain.Lesson, error) {
	lesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return nil, err
	}

	// Завершить урок может преподаватель этого урока или администратор школы
	if callerRole != domain.RoleOwner && lesson.TeacherID != callerID {
		return nil, domain.ErrUnauthorizedLessonAction
	}

	if err := lesson.Complete(); err != nil {
		return nil, err
	}

	if err := s.lessonRepo.Update(ctx, lesson); err != nil {
		return nil, fmt.Errorf("update completed lesson: %w", err)
	}

	return lesson, nil
}

func (s *Service) CancelLesson(ctx context.Context, lessonID, callerID uuid.UUID, callerRole domain.Role, reason string) (*domain.Lesson, error) {
	lesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return nil, err
	}

	// Отменить урок может ученик, преподаватель урока или владелец
	if callerRole != domain.RoleOwner && lesson.TeacherID != callerID && lesson.StudentID != callerID {
		return nil, domain.ErrUnauthorizedLessonAction
	}

	cancelRole := callerRole
	if callerRole != domain.RoleStudent && callerRole != domain.RoleTeacher {
		if callerID == lesson.StudentID {
			cancelRole = domain.RoleStudent
		} else {
			cancelRole = domain.RoleTeacher
		}
	}

	if err := lesson.Cancel(cancelRole, reason); err != nil {
		return nil, err
	}

	if err := s.lessonRepo.Update(ctx, lesson); err != nil {
		return nil, fmt.Errorf("update cancelled lesson: %w", err)
	}

	return lesson, nil
}

func (s *Service) ListLessons(ctx context.Context, filter LessonFilter) ([]*domain.Lesson, error) {
	return s.lessonRepo.List(ctx, filter)
}

func (s *Service) GetLesson(ctx context.Context, lessonID uuid.UUID) (*domain.Lesson, error) {
	return s.lessonRepo.GetByID(ctx, lessonID)
}
