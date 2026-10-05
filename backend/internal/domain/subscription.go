package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSubscriptionNotFound      = errors.New("subscription not found")
	ErrInvalidSubscriptionFormat = errors.New("invalid subscription format, must be individual, pair, or group")
)

type SubscriptionFormat string

const (
	SubscriptionFormatIndividual SubscriptionFormat = "individual"
	SubscriptionFormatPair       SubscriptionFormat = "pair"
	SubscriptionFormatGroup      SubscriptionFormat = "group"
)

func (f SubscriptionFormat) IsValid() bool {
	return f == SubscriptionFormatIndividual || f == SubscriptionFormatPair || f == SubscriptionFormatGroup
}

// SubscriptionType оставлен для обратной совместимости.
type SubscriptionType = SubscriptionFormat

const (
	SubscriptionTypeLessons SubscriptionType = "lessons"
	SubscriptionTypeHours   SubscriptionType = "hours"
)

type ClientSubscription struct {
	ID        uuid.UUID          `json:"id" db:"id"`
	ClientID  uuid.UUID          `json:"client_id" db:"client_id"`
	Format    SubscriptionFormat `json:"format" db:"format"`
	Balance   float64            `json:"balance" db:"balance"`
	CreatedAt time.Time          `json:"created_at" db:"created_at"`
}

type ClientBalanceAdjustment struct {
	ID         uuid.UUID          `json:"id" db:"id"`
	ClientID   uuid.UUID          `json:"client_id" db:"client_id"`
	TeacherID  uuid.UUID          `json:"teacher_id" db:"teacher_id"`
	Format     SubscriptionFormat `json:"format" db:"format"`
	DeltaHours float64            `json:"delta_hours" db:"delta_hours"`
	Reason     string             `json:"reason" db:"reason"`
	CreatedAt  time.Time          `json:"created_at" db:"created_at"`
}

