package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrJournalNotFound         = errors.New("lesson journal not found")
	ErrInvalidPerformanceScore = errors.New("performance score must be between 1 and 5")
	ErrEmptyTopic              = errors.New("topic cannot be empty")
)

// LessonJournal представляет дидактический отчет по проведенному уроку (1:1 к Lesson).
type LessonJournal struct {
	ID               uuid.UUID `json:"id" db:"id"`
	LessonID         uuid.UUID `json:"lesson_id" db:"lesson_id"`
	ClientID         uuid.UUID `json:"client_id" db:"client_id"`
	TeacherID        uuid.UUID `json:"teacher_id" db:"teacher_id"`
	Topic            string    `json:"topic" db:"topic"`
	Notes            string    `json:"notes" db:"notes"`
	PerformanceScore *int      `json:"performance_score,omitempty" db:"performance_score"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// NewLessonJournal создает новый отчет по уроку с валидацией.
func NewLessonJournal(
	lessonID uuid.UUID,
	clientID uuid.UUID,
	teacherID uuid.UUID,
	topic string,
	notes string,
	score *int,
) (*LessonJournal, error) {
	journal := &LessonJournal{
		ID:               uuid.New(),
		LessonID:         lessonID,
		ClientID:         clientID,
		TeacherID:        teacherID,
		Topic:            strings.TrimSpace(topic),
		Notes:            strings.TrimSpace(notes),
		PerformanceScore: score,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	if err := journal.Validate(); err != nil {
		return nil, err
	}

	return journal, nil
}

// Validate проверяет корректность полей журнала урока.
func (j *LessonJournal) Validate() error {
	if j.LessonID == uuid.Nil {
		return errors.New("lesson_id is required")
	}
	if j.ClientID == uuid.Nil {
		return errors.New("client_id is required")
	}
	if j.TeacherID == uuid.Nil {
		return errors.New("teacher_id is required")
	}
	if strings.TrimSpace(j.Topic) == "" {
		return ErrEmptyTopic
	}
	if j.PerformanceScore != nil {
		if *j.PerformanceScore < 1 || *j.PerformanceScore > 5 {
			return ErrInvalidPerformanceScore
		}
	}
	return nil
}

// Update обновляет поля отчета по уроку.
func (j *LessonJournal) Update(topic string, notes string, score *int) error {
	trimmedTopic := strings.TrimSpace(topic)
	if trimmedTopic == "" {
		return ErrEmptyTopic
	}
	if score != nil && (*score < 1 || *score > 5) {
		return ErrInvalidPerformanceScore
	}

	j.Topic = trimmedTopic
	j.Notes = strings.TrimSpace(notes)
	j.PerformanceScore = score
	j.UpdatedAt = time.Now().UTC()
	return nil
}
