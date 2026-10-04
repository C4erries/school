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
	StatusPendingConfirmation LessonStatus = "pending_confirmation"
	StatusConfirmed           LessonStatus = "confirmed"
	StatusCompleted           LessonStatus = "completed"
	StatusCancelledByTeacher  LessonStatus = "cancelled_by_teacher"
	StatusCancelledByStudent  LessonStatus = "cancelled_by_student"
	StatusDeclined            LessonStatus = "declined"
	StatusNoShow              LessonStatus = "no_show"
)

func (s LessonStatus) String() string {
	return string(s)
}

func (s LessonStatus) IsValid() bool {
	switch s {
	case StatusPendingConfirmation,
		StatusConfirmed,
		StatusCompleted,
		StatusCancelledByTeacher,
		StatusCancelledByStudent,
		StatusDeclined,
		StatusNoShow:
		return true
	default:
		return false
	}
}

// Lesson представляет сущность урока в домене.
type Lesson struct {
	ID           uuid.UUID
	TeacherID    uuid.UUID
	StudentID    uuid.UUID
	ClassroomID  *uuid.UUID
	StartTime    time.Time
	EndTime      time.Time
	Format       LessonFormat
	LocationOrURL string
	Status       LessonStatus
	Notes        string
	CancelReason string
	CreatedAt    time.Time
	UpdatedAt    time.Time
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
	ErrLessonAlreadyFinished       = errors.New("lesson is already completed, cancelled or declined")
)

// ValidateLessonTimes проверяет корректность временных границ урока.
func ValidateLessonTimes(start, end time.Time) error {
	if !end.After(start) {
		return ErrInvalidTimeRange
	}
	return nil
}

// CanAccept проверяет, можно ли принять урок (только из pending_confirmation).
func (l *Lesson) CanAccept() bool {
	return l.Status == StatusPendingConfirmation
}

// Accept переводит урок в статус confirmed.
func (l *Lesson) Accept() error {
	if !l.CanAccept() {
		return ErrInvalidLessonStatus
	}
	l.Status = StatusConfirmed
	l.UpdatedAt = time.Now().UTC()
	return nil
}

// CanDecline проверяет, можно ли отклонить урок (только из pending_confirmation).
func (l *Lesson) CanDecline() bool {
	return l.Status == StatusPendingConfirmation
}

// Decline переводит урок в статус declined с указанием причины.
func (l *Lesson) Decline(reason string) error {
	if !l.CanDecline() {
		return ErrInvalidLessonStatus
	}
	l.Status = StatusDeclined
	l.CancelReason = strings.TrimSpace(reason)
	l.UpdatedAt = time.Now().UTC()
	return nil
}

// CanComplete проверяет, можно ли завершить урок (только из confirmed).
func (l *Lesson) CanComplete() bool {
	return l.Status == StatusConfirmed
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

// CanCancel проверяет, можно ли отменить урок (из pending_confirmation или confirmed).
func (l *Lesson) CanCancel() bool {
	return l.Status == StatusPendingConfirmation || l.Status == StatusConfirmed
}

// Cancel отменяет урок пользователем с определенной ролью (преподаватель/ученик/админ).
func (l *Lesson) Cancel(role Role, reason string) error {
	if !l.CanCancel() {
		return ErrInvalidLessonStatus
	}

	if role == RoleStudent {
		l.Status = StatusCancelledByStudent
	} else {
		// Преподаватель или владелец/администратор
		l.Status = StatusCancelledByTeacher
	}

	l.CancelReason = strings.TrimSpace(reason)
	l.UpdatedAt = time.Now().UTC()
	return nil
}

// CanMarkNoShow проверяет, можно ли отметить неявку (только из confirmed).
func (l *Lesson) CanMarkNoShow() bool {
	return l.Status == StatusConfirmed
}

// MarkNoShow отмечает неявку ученика на подтвержденный урок.
func (l *Lesson) MarkNoShow(reason string) error {
	if !l.CanMarkNoShow() {
		return ErrInvalidLessonStatus
	}
	l.Status = StatusNoShow
	l.CancelReason = strings.TrimSpace(reason)
	l.UpdatedAt = time.Now().UTC()
	return nil
}
