package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTagNotFound           = errors.New("tag not found")
	ErrInvalidTagName        = errors.New("tag name is required")
	ErrInvalidSchoolPercent  = errors.New("school percent must be between 0 and 100")
	ErrUnauthorizedTagAction = errors.New("unauthorized action for this tag")
)

type Tag struct {
	ID            uuid.UUID `json:"id" db:"id"`
	TeacherID     uuid.UUID `json:"teacher_id" db:"teacher_id"`
	Name          string    `json:"name" db:"name"`
	SchoolPercent int       `json:"school_percent" db:"school_percent"`
	Color         string    `json:"color" db:"color"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

func ValidateTagInput(name string, percent int) (string, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return "", ErrInvalidTagName
	}
	if percent < 0 || percent > 100 {
		return "", ErrInvalidSchoolPercent
	}
	return trimmedName, nil
}
