package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

type SubscriptionType string

const (
	SubscriptionTypeLessons SubscriptionType = "lessons"
	SubscriptionTypeHours   SubscriptionType = "hours"
)

type ClientSubscription struct {
	ID        uuid.UUID        `json:"id" db:"id"`
	ClientID  uuid.UUID        `json:"client_id" db:"client_id"`
	Type      SubscriptionType `json:"type" db:"type"`
	Balance   float64          `json:"balance" db:"balance"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
}
