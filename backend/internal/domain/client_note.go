package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrClientNoteNotFound      = errors.New("client note not found")
	ErrEmptyClientNoteContent = errors.New("client note content cannot be empty")
)

// ClientNote представляет быструю свободную заметку репетитора по ученику.
type ClientNote struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ClientID  uuid.UUID `json:"client_id" db:"client_id"`
	TeacherID uuid.UUID `json:"teacher_id" db:"teacher_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NewClientNote создает новую свободную заметку по ученику с валидацией.
func NewClientNote(clientID, teacherID uuid.UUID, content string) (*ClientNote, error) {
	note := &ClientNote{
		ID:        uuid.New(),
		ClientID:  clientID,
		TeacherID: teacherID,
		Content:   strings.TrimSpace(content),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := note.Validate(); err != nil {
		return nil, err
	}

	return note, nil
}

// Validate проверяет корректность полей свободной заметки.
func (n *ClientNote) Validate() error {
	if n.ClientID == uuid.Nil {
		return errors.New("client_id is required")
	}
	if n.TeacherID == uuid.Nil {
		return errors.New("teacher_id is required")
	}
	if strings.TrimSpace(n.Content) == "" {
		return ErrEmptyClientNoteContent
	}
	return nil
}

// Update обновляет текст заметки с валидацией.
func (n *ClientNote) Update(content string) error {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return ErrEmptyClientNoteContent
	}
	n.Content = trimmed
	n.UpdatedAt = time.Now().UTC()
	return nil
}

