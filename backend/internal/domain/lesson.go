package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// LessonFormat определяет формат проведения занятия.
type LessonFormat string

const (
	FormatOnline  LessonFormat = "online"
	FormatOffline LessonFormat = "offline"
)

func (f LessonFormat) String() string {
	return string(f)
}

func (f LessonFormat) IsValid() bool {
	return f == FormatOnline || f == FormatOffline
}

// LessonStatus определяет статус жизненного цикла занятия.
type LessonStatus string

const (
	StatusScheduled LessonStatus = "scheduled"
	StatusCompleted LessonStatus = "completed"
	StatusCancelled LessonStatus = "cancelled"
)

func (s LessonStatus) String() string {
	return string(s)
}

func (s LessonStatus) IsValid() bool {
	switch s {
	case StatusScheduled,
		StatusCompleted,
		StatusCancelled:
		return true
	default:
		return false
	}
}

// Lesson представляет сущность урока в домене.
type Lesson struct {
	ID            uuid.UUID
	TeacherID     uuid.UUID
	ClientID      uuid.UUID
	ClassroomID   *uuid.UUID
	StartTime     time.Time
	EndTime       time.Time
	Format        LessonFormat
	LocationOrURL string
	Status        LessonStatus
	Notes         string
	CancelReason  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Ошибки работы с уроками.
var (
	ErrLessonNotFound              = errors.New("lesson not found")
	ErrInvalidTimeRange            = errors.New("lesson end time must be after start time")
	ErrLessonInPast                = errors.New("cannot schedule lesson in the past")
	ErrClassroomCollision          = errors.New("classroom is already booked by another teacher for this time")
	ErrInvalidLessonStatus         = errors.New("invalid lesson status transition")
	ErrUnauthorizedLessonAction    = errors.New("unauthorized action for this lesson")
	ErrClassroomRequiredForOffline = errors.New("classroom is required for offline lesson")
	ErrInvalidLessonFormat         = errors.New("invalid lesson format, must be online or offline")
	ErrLessonAlreadyFinished       = errors.New("lesson is already completed or cancelled")
)

// ValidateLessonTimes проверяет корректность временных границ урока.
func ValidateLessonTimes(start, end time.Time) error {
	if !end.After(start) {
		return ErrInvalidTimeRange
	}
	return nil
}

// CanComplete проверяет, можно ли завершить урок (только из scheduled).
func (l *Lesson) CanComplete() bool {
	return l.Status == StatusScheduled
}

// Complete переводит подтвержденный урок в статус completed.
func (l *Lesson) Complete() error {
	if !l.CanComplete() {
		return ErrInvalidLessonStatus
	}
	l.Status = StatusCompleted
	l.UpdatedAt = time.Now().UTC()
	return nil
}

// CanCancel проверяет, можно ли отменить урок (из scheduled).
func (l *Lesson) CanCancel() bool {
	return l.Status == StatusScheduled
}

// Cancel отменяет урок пользователем.
func (l *Lesson) Cancel(reason string) error {
	if !l.CanCancel() {
		return ErrInvalidLessonStatus
	}

	l.Status = StatusCancelled
	l.CancelReason = strings.TrimSpace(reason)
	l.UpdatedAt = time.Now().UTC()
	return nil
}
