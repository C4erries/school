package domain

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Role определяет роль пользователя в системе школы.
type Role string

const (
	RoleStudent   Role = "student"
	RoleTeacher   Role = "teacher"
	RoleAssistant Role = "assistant"
	RoleOwner     Role = "owner"
)

func (r Role) String() string {
	return string(r)
}

func (r Role) IsValid() bool {
	switch r {
	case RoleStudent, RoleTeacher, RoleAssistant, RoleOwner:
		return true
	default:
		return false
	}
}

// User представляет сущность пользователя в домене.
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FullName     string
	Phone        *string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Ошибки валидации и существования пользователя.
var (
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrInvalidPassword    = errors.New("password must be at least 6 characters")
	ErrInvalidFullName    = errors.New("full name must be at least 2 characters")
	ErrInvalidRole        = errors.New("invalid user role")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrSessionExpired     = errors.New("session expired or revoked")
)

// ValidateEmail проверяет корректность формата email.
func ValidateEmail(email string) (string, error) {
	clean := strings.ToLower(strings.TrimSpace(email))
	addr, err := mail.ParseAddress(clean)
	if err != nil || addr.Address != clean || clean == "" {
		return "", ErrInvalidEmail
	}
	return clean, nil
}

// ValidatePassword проверяет минимальную стойкость пароля.
func ValidatePassword(password string) error {
	if len(password) < 6 {
		return ErrInvalidPassword
	}
	return nil
}

// ValidateFullName проверяет имя пользователя.
func ValidateFullName(fullName string) (string, error) {
	clean := strings.TrimSpace(fullName)
	if len(clean) < 2 {
		return "", ErrInvalidFullName
	}
	return clean, nil
}
