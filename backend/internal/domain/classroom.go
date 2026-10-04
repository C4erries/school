package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Classroom представляет учебный кабинет школы.
type Classroom struct {
	ID          uuid.UUID
	Name        string
	Capacity    int
	Color       string
	Description string
	CreatedAt   time.Time
}

// Ошибки кабинета.
var (
	ErrClassroomNotFound        = errors.New("classroom not found")
	ErrInvalidClassroomName     = errors.New("classroom name must be at least 2 characters")
	ErrInvalidClassroomCapacity = errors.New("classroom capacity must be greater than 0")
)

// ValidateClassroomInput проверяет корректность входных данных для кабинета.
func ValidateClassroomInput(name string, capacity int) (string, error) {
	cleanName := strings.TrimSpace(name)
	if len(cleanName) < 2 {
		return "", ErrInvalidClassroomName
	}
	if capacity <= 0 {
		return "", ErrInvalidClassroomCapacity
	}
	return cleanName, nil
}
