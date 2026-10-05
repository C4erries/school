package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrClientNotFound = errors.New("client not found")
)

type Client struct {
	ID               uuid.UUID `json:"id" db:"id"`
	TeacherID        uuid.UUID `json:"teacher_id" db:"teacher_id"`
	Name             string    `json:"name" db:"name"`
	Phone            *string   `json:"phone" db:"phone"`
	BaseRate         float64   `json:"base_rate" db:"base_rate"`
	SchoolPercentTag int       `json:"school_percent_tag" db:"school_percent_tag"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}
