package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPaymentNotFound           = errors.New("payment not found")
	ErrInvalidPaymentAmount      = errors.New("payment amount must be greater than zero")
	ErrInvalidPaymentHours       = errors.New("payment hours must be greater than zero")
	ErrInvalidPaymentMethod      = errors.New("invalid payment method")
	ErrInvalidPaymentFormat      = errors.New("invalid payment format, must be individual, pair, or group")
	ErrUnauthorizedPaymentAction = errors.New("unauthorized action for this payment")
)

// PaymentMethod определяет способ оплаты.
type PaymentMethod string

const (
	PaymentMethodTransfer PaymentMethod = "transfer"
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodCard     PaymentMethod = "card"
	PaymentMethodOther    PaymentMethod = "other"
)

func (m PaymentMethod) String() string {
	return string(m)
}

func (m PaymentMethod) IsValid() bool {
	switch m {
	case PaymentMethodTransfer, PaymentMethodCash, PaymentMethodCard, PaymentMethodOther:
		return true
	default:
		return false
	}
}

// Payment представляет факт внесения оплаты учеником за абонемент.
type Payment struct {
	ID            uuid.UUID          `json:"id" db:"id"`
	TeacherID     uuid.UUID          `json:"teacher_id" db:"teacher_id"`
	ClientID      uuid.UUID          `json:"client_id" db:"client_id"`
	Amount        float64            `json:"amount" db:"amount"`
	Hours         float64            `json:"hours" db:"hours"`
	Format        SubscriptionFormat `json:"format" db:"format"`
	PaymentMethod PaymentMethod      `json:"payment_method" db:"payment_method"`
	PaidAt        time.Time          `json:"paid_at" db:"paid_at"`
	Notes         string             `json:"notes" db:"notes"`
	CreatedAt     time.Time          `json:"created_at" db:"created_at"`
}

// Validate проверяет валидность полей платежа.
func (p *Payment) Validate() error {
	if p.Amount <= 0 {
		return ErrInvalidPaymentAmount
	}
	if p.Hours <= 0 {
		return ErrInvalidPaymentHours
	}
	if !p.Format.IsValid() {
		return ErrInvalidPaymentFormat
	}
	if !p.PaymentMethod.IsValid() {
		return ErrInvalidPaymentMethod
	}
	return nil
}
