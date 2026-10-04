package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// TeacherStudent представляет связь прикрепления ученика к преподавателю.
type TeacherStudent struct {
	ID        uuid.UUID
	TeacherID uuid.UUID
	StudentID uuid.UUID
	CreatedAt time.Time
}

// Ошибки прикрепления ученика к преподавателю.
var (
	ErrTeacherStudentAlreadyExists = errors.New("student is already assigned to this teacher")
	ErrTeacherStudentNotFound      = errors.New("teacher-student assignment not found")
	ErrCannotAssignSelf            = errors.New("cannot assign teacher to themselves")
	ErrInvalidTeacherRole          = errors.New("assigned teacher must have role 'teacher'")
	ErrInvalidStudentRole          = errors.New("assigned student must have role 'student'")
	ErrStudentNotAssignedToTeacher = errors.New("student is not assigned to this teacher")
)
