package schedule

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
)

// LessonFilter параметры выборки списка уроков.
type LessonFilter struct {
	TeacherID   *uuid.UUID
	ClientID    *uuid.UUID
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

// ClientRepository определяет контракт хранилища клиентов.
type ClientRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error)
}

// SubscriptionRepository определяет контракт хранилища абонементов.
type SubscriptionRepository interface {
	GetByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.ClientSubscription, error)
	Update(ctx context.Context, sub *domain.ClientSubscription) error
	Create(ctx context.Context, sub *domain.ClientSubscription) error
}

// LessonRepository определяет контракт хранилища занятий.
type LessonRepository interface {
	Create(ctx context.Context, l *domain.Lesson) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Lesson, error)
	Update(ctx context.Context, l *domain.Lesson) error
	HasClassroomCollision(ctx context.Context, classroomID uuid.UUID, teacherID uuid.UUID, startTime, endTime time.Time, excludeLessonID *uuid.UUID) (bool, error)
	List(ctx context.Context, filter LessonFilter) ([]*domain.Lesson, error)
}

// LessonSeriesRepository определяет контракт хранилища регулярных серий занятий.
type LessonSeriesRepository interface {
	Create(ctx context.Context, s *domain.LessonSeries) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.LessonSeries, error)
	Update(ctx context.Context, s *domain.LessonSeries) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.LessonSeries, error)
}

type ScheduleLessonInput struct {
	TeacherID     uuid.UUID
	ClientID      uuid.UUID
	Title         string
	ClassroomID   *uuid.UUID
	StartTime     time.Time
	EndTime       time.Time
	Format        domain.LessonFormat
	LocationOrURL string
	Notes         string
}

// CreateLessonInput алиас для ScheduleLessonInput.
type CreateLessonInput = ScheduleLessonInput

type UpdateLessonInput struct {
	LessonID       uuid.UUID
	CallerID       uuid.UUID
	CallerRole     domain.Role
	Title          *string
	ClientID       *uuid.UUID
	ClassroomID    *uuid.UUID
	ClearClassroom bool
	StartTime      *time.Time
	EndTime        *time.Time
	Format         *domain.LessonFormat
	LocationOrURL  *string
	Notes          *string
	CancelReason   *string
}

// Service реализует бизнес-логику расписания, аудиторий и проведения уроков.
type Service struct {
	classroomRepo ClassroomRepository
	lessonRepo    LessonRepository
	clientRepo    ClientRepository
	subRepo       SubscriptionRepository
	seriesRepo    LessonSeriesRepository
}

func NewService(
	classroomRepo ClassroomRepository,
	lessonRepo LessonRepository,
	clientRepo ClientRepository,
	subRepo SubscriptionRepository,
	seriesRepo ...LessonSeriesRepository,
) *Service {
	var sRepo LessonSeriesRepository
	if len(seriesRepo) > 0 {
		sRepo = seriesRepo[0]
	}
	return &Service{
		classroomRepo: classroomRepo,
		lessonRepo:    lessonRepo,
		clientRepo:    clientRepo,
		subRepo:       subRepo,
		seriesRepo:    sRepo,
	}
}

// --- Уроки (Lessons) ---

func (s *Service) ScheduleLesson(ctx context.Context, input ScheduleLessonInput) (*domain.Lesson, error) {
	if !input.Format.IsValid() {
		return nil, domain.ErrInvalidLessonFormat
	}

	if err := domain.ValidateLessonTimes(input.StartTime, input.EndTime); err != nil {
		return nil, err
	}

	client, err := s.clientRepo.GetByID(ctx, input.ClientID)
	if err != nil {
		return nil, fmt.Errorf("get client: %w", err)
	}

	if client.TeacherID != input.TeacherID {
		return nil, domain.ErrUnauthorizedLessonAction
	}

	var classroomID *uuid.UUID
	if input.ClassroomID != nil && *input.ClassroomID != uuid.Nil {
		if _, err := s.classroomRepo.GetByID(ctx, *input.ClassroomID); err != nil {
			return nil, err
		}

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

	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "Занятие"
	}

	now := time.Now().UTC()
	lesson := &domain.Lesson{
		ID:            uuid.New(),
		TeacherID:     input.TeacherID,
		ClientID:      input.ClientID,
		Title:         title,
		ClassroomID:   classroomID,
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		Format:        input.Format,
		LocationOrURL: strings.TrimSpace(input.LocationOrURL),
		Status:        domain.StatusScheduled,
		Notes:         strings.TrimSpace(input.Notes),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.lessonRepo.Create(ctx, lesson); err != nil {
		return nil, fmt.Errorf("create lesson: %w", err)
	}

	return lesson, nil
}

func (s *Service) CompleteLesson(ctx context.Context, lessonID, callerID uuid.UUID, callerRole domain.Role) (*domain.Lesson, error) {
	lesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		if errors.Is(err, domain.ErrLessonNotFound) {
			series, slot, vErr := s.FindVirtualLesson(ctx, lessonID, &callerID)
			if vErr == nil && series != nil && slot != nil {
				return s.CompleteRecurringLesson(ctx, lessonID, &series.ID, &slot.OriginalStartTime, callerID, callerRole)
			}
		}
		return nil, err
	}

	if callerRole != domain.RoleOwner && lesson.TeacherID != callerID {
		return nil, domain.ErrUnauthorizedLessonAction
	}

	if err := lesson.Complete(); err != nil {
		return nil, err
	}

	if s.subRepo != nil {
		durationHours := lesson.EndTime.Sub(lesson.StartTime).Seconds() / 3600.0
		targetFormat := domain.SubscriptionFormat(lesson.Format)

		subs, err := s.subRepo.GetByClientID(ctx, lesson.ClientID)
		if err == nil {
			var matchingSub *domain.ClientSubscription
			for _, sub := range subs {
				if sub.Format == targetFormat {
					matchingSub = sub
					break
				}
			}

			if matchingSub != nil {
				matchingSub.Balance -= durationHours
				_ = s.subRepo.Update(ctx, matchingSub)
			} else {
				newSub := &domain.ClientSubscription{
					ID:        uuid.New(),
					ClientID:  lesson.ClientID,
					Format:    targetFormat,
					Balance:   -durationHours,
					CreatedAt: time.Now().UTC(),
				}
				_ = s.subRepo.Create(ctx, newSub)
			}
		}
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

	if callerRole != domain.RoleOwner && lesson.TeacherID != callerID {
		return nil, domain.ErrUnauthorizedLessonAction
	}

	if err := lesson.Cancel(reason); err != nil {
		return nil, err
	}

	if err := s.lessonRepo.Update(ctx, lesson); err != nil {
		return nil, fmt.Errorf("update cancelled lesson: %w", err)
	}

	return lesson, nil
}

func (s *Service) UpdateLesson(ctx context.Context, input UpdateLessonInput) (*domain.Lesson, error) {
	lesson, err := s.lessonRepo.GetByID(ctx, input.LessonID)
	if err != nil {
		return nil, err
	}

	if input.CallerRole != domain.RoleOwner && lesson.TeacherID != input.CallerID {
		return nil, domain.ErrUnauthorizedLessonAction
	}

	startTime := lesson.StartTime
	endTime := lesson.EndTime
	if input.StartTime != nil {
		startTime = *input.StartTime
	}
	if input.EndTime != nil {
		endTime = *input.EndTime
	}

	if err := domain.ValidateLessonTimes(startTime, endTime); err != nil {
		return nil, err
	}
	lesson.StartTime = startTime
	lesson.EndTime = endTime

	if input.Format != nil {
		if !input.Format.IsValid() {
			return nil, domain.ErrInvalidLessonFormat
		}
		lesson.Format = *input.Format
	}

	if input.ClientID != nil {
		client, err := s.clientRepo.GetByID(ctx, *input.ClientID)
		if err != nil {
			return nil, fmt.Errorf("get client: %w", err)
		}
		if client.TeacherID != lesson.TeacherID {
			return nil, domain.ErrUnauthorizedLessonAction
		}
		lesson.ClientID = *input.ClientID
	}

	if input.ClearClassroom {
		lesson.ClassroomID = nil
	} else if input.ClassroomID != nil {
		if *input.ClassroomID == uuid.Nil {
			lesson.ClassroomID = nil
		} else {
			if _, err := s.classroomRepo.GetByID(ctx, *input.ClassroomID); err != nil {
				return nil, err
			}
			lesson.ClassroomID = input.ClassroomID
		}
	}

	if lesson.ClassroomID != nil && *lesson.ClassroomID != uuid.Nil {
		collision, err := s.lessonRepo.HasClassroomCollision(
			ctx,
			*lesson.ClassroomID,
			lesson.TeacherID,
			lesson.StartTime,
			lesson.EndTime,
			&lesson.ID,
		)
		if err != nil {
			return nil, fmt.Errorf("check classroom collision: %w", err)
		}
		if collision {
			return nil, domain.ErrClassroomCollision
		}
	}

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			title = "Занятие"
		}
		lesson.Title = title
	}

	if input.LocationOrURL != nil {
		lesson.LocationOrURL = strings.TrimSpace(*input.LocationOrURL)
	}
	if lesson.ClassroomID != nil && *lesson.ClassroomID != uuid.Nil && (input.LocationOrURL == nil || strings.TrimSpace(*input.LocationOrURL) == "") {
		lesson.LocationOrURL = ""
	}

	if input.Notes != nil {
		lesson.Notes = strings.TrimSpace(*input.Notes)
	}
	if input.CancelReason != nil {
		lesson.CancelReason = strings.TrimSpace(*input.CancelReason)
	}

	lesson.UpdatedAt = time.Now().UTC()

	if err := s.lessonRepo.Update(ctx, lesson); err != nil {
		return nil, fmt.Errorf("update lesson: %w", err)
	}

	return lesson, nil
}

func (s *Service) ListLessons(ctx context.Context, filter LessonFilter) ([]*domain.Lesson, error) {
	return s.ListLessonsWithRecurring(ctx, filter)
}

func (s *Service) GetLesson(ctx context.Context, lessonID uuid.UUID) (*domain.Lesson, error) {
	return s.lessonRepo.GetByID(ctx, lessonID)
}

// GetUpcomingLesson возвращает ближайший запланированный урок ученика (начиная от текущего момента time.Now().UTC()).
func (s *Service) GetUpcomingLesson(ctx context.Context, clientID, teacherID uuid.UUID) (*domain.Lesson, error) {
	now := time.Now().UTC()
	to := now.AddDate(1, 0, 0) // горизонт 1 год вперед
	status := domain.StatusScheduled
	filter := LessonFilter{
		TeacherID: &teacherID,
		ClientID:  &clientID,
		Status:    &status,
		From:      &now,
		To:        &to,
	}

	lessons, err := s.ListLessonsWithRecurring(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list upcoming lessons: %w", err)
	}

	for _, l := range lessons {
		if !l.StartTime.Before(now) && l.Status == domain.StatusScheduled {
			return l, nil
		}
	}

	return nil, nil
}
