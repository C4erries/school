package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// HomeworkStatus определяет текущее состояние домашнего задания.
type HomeworkStatus string

const (
	StatusHomeworkAssigned  HomeworkStatus = "assigned"
	StatusHomeworkCompleted HomeworkStatus = "completed"
	StatusHomeworkNotDone   HomeworkStatus = "not_done"
)

func (s HomeworkStatus) String() string {
	return string(s)
}

func (s HomeworkStatus) IsValid() bool {
	switch s {
	case StatusHomeworkAssigned, StatusHomeworkCompleted, StatusHomeworkNotDone:
		return true
	default:
		return false
	}
}

var (
	ErrHomeworkNotFound      = errors.New("homework assignment not found")
	ErrInvalidHomeworkStatus = errors.New("invalid homework status")
	ErrEmptyHomeworkTitle    = errors.New("homework title cannot be empty")
)

// HomeworkAssignment представляет домашнее задание, выданное ученику.
type HomeworkAssignment struct {
	ID               uuid.UUID      `json:"id" db:"id"`
	ClientID         uuid.UUID      `json:"client_id" db:"client_id"`
	TeacherID        uuid.UUID      `json:"teacher_id" db:"teacher_id"`
	AssignedLessonID *uuid.UUID     `json:"assigned_lesson_id,omitempty" db:"assigned_lesson_id"`
	Title            string         `json:"title" db:"title"`
	Description      string         `json:"description" db:"description"`
	DueDate          *time.Time     `json:"due_date,omitempty" db:"due_date"`
	Status           HomeworkStatus `json:"status" db:"status"`
	ReviewNotes      string         `json:"review_notes" db:"review_notes"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at" db:"updated_at"`
}

// NewHomeworkAssignment создает новое домашнее задание в статусе assigned.
func NewHomeworkAssignment(
	clientID uuid.UUID,
	teacherID uuid.UUID,
	assignedLessonID *uuid.UUID,
	title string,
	description string,
	dueDate *time.Time,
) (*HomeworkAssignment, error) {
	hw := &HomeworkAssignment{
		ID:               uuid.New(),
		ClientID:         clientID,
		TeacherID:        teacherID,
		AssignedLessonID: assignedLessonID,
		Title:            strings.TrimSpace(title),
		Description:      strings.TrimSpace(description),
		DueDate:          dueDate,
		Status:           StatusHomeworkAssigned,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	if err := hw.Validate(); err != nil {
		return nil, err
	}

	return hw, nil
}

// Validate проверяет корректность полей домашнего задания.
func (h *HomeworkAssignment) Validate() error {
	if h.ClientID == uuid.Nil {
		return errors.New("client_id is required")
	}
	if h.TeacherID == uuid.Nil {
		return errors.New("teacher_id is required")
	}
	if strings.TrimSpace(h.Title) == "" {
		return ErrEmptyHomeworkTitle
	}
	if !h.Status.IsValid() {
		return ErrInvalidHomeworkStatus
	}
	return nil
}

// UpdateStatus обновляет статус выполнения и рецензию преподавателя.
func (h *HomeworkAssignment) UpdateStatus(status HomeworkStatus, reviewNotes string) error {
	if !status.IsValid() {
		return ErrInvalidHomeworkStatus
	}
	h.Status = status
	h.ReviewNotes = strings.TrimSpace(reviewNotes)
	h.UpdatedAt = time.Now().UTC()
	return nil
}

// MarkCompleted переводит задание в статус выполненного с опциональным комментарием.
func (h *HomeworkAssignment) MarkCompleted(reviewNotes string) error {
	return h.UpdateStatus(StatusHomeworkCompleted, reviewNotes)
}

// MarkNotDone переводит задание в статус невыполненного с опциональным комментарием.
func (h *HomeworkAssignment) MarkNotDone(reviewNotes string) error {
	return h.UpdateStatus(StatusHomeworkNotDone, reviewNotes)
}
