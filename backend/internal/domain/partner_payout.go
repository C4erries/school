package domain

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPartnerPayoutNotFound    = errors.New("partner payout not found")
	ErrInvalidPeriodMonth       = errors.New("invalid period month, must be in YYYY-MM format")
	ErrInvalidCommissionAmount  = errors.New("commission amount must be non-negative")
	ErrInvalidGrossAmount       = errors.New("gross amount must be non-negative")
	ErrUnauthorizedPayoutAction = errors.New("unauthorized action for this partner payout")
)

var periodMonthRegex = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)

// PartnerPayout представляет зафиксированную выплату комиссии партнерской школе.
type PartnerPayout struct {
	ID               uuid.UUID `json:"id" db:"id"`
	TeacherID        uuid.UUID `json:"teacher_id" db:"teacher_id"`
	TagID            uuid.UUID `json:"tag_id" db:"tag_id"`
	PeriodMonth      string    `json:"period_month" db:"period_month"` // Формат YYYY-MM
	GrossAmount      float64   `json:"gross_amount" db:"gross_amount"`
	CommissionAmount float64   `json:"commission_amount" db:"commission_amount"`
	PaidAt           time.Time `json:"paid_at" db:"paid_at"`
	Notes            string    `json:"notes" db:"notes"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// ValidatePeriodMonth проверяет корректность формата строки YYYY-MM.
func ValidatePeriodMonth(period string) bool {
	return periodMonthRegex.MatchString(period)
}

// Validate проверяет валидность полей выплаты.
func (p *PartnerPayout) Validate() error {
	if !ValidatePeriodMonth(p.PeriodMonth) {
		return ErrInvalidPeriodMonth
	}
	if p.GrossAmount < 0 {
		return ErrInvalidGrossAmount
	}
	if p.CommissionAmount < 0 {
		return ErrInvalidCommissionAmount
	}
	return nil
}
