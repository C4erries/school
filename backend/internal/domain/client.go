package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrClientNotFound = errors.New("client not found")
)

type ClientBalances struct {
	IndividualHours float64 `json:"individual_hours"`
	PairHours       float64 `json:"pair_hours"`
	GroupHours      float64 `json:"group_hours"`
	TotalHours      float64 `json:"total_hours"`
}

type Client struct {
	ID               uuid.UUID      `json:"id" db:"id"`
	TeacherID        uuid.UUID      `json:"teacher_id" db:"teacher_id"`
	Name             string         `json:"name" db:"name"`
	Phone            *string        `json:"phone" db:"phone"`
	BaseRate         float64        `json:"base_rate" db:"base_rate"`
	RateIndividual   float64        `json:"rate_individual" db:"rate_individual"`
	RatePair         *float64       `json:"rate_pair" db:"rate_pair"`
	RateGroup        *float64       `json:"rate_group" db:"rate_group"`
	SchoolPercentTag int            `json:"school_percent_tag" db:"school_percent_tag"`
	Tags             []Tag          `json:"tags"`
	Balances         ClientBalances `json:"balances"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
}

// RateForFormat возвращает ставку для заданного формата занятия.
func (c *Client) RateForFormat(format LessonFormat) float64 {
	switch format {
	case FormatPair:
		if c.RatePair != nil && *c.RatePair > 0 {
			return *c.RatePair
		}
	case FormatGroup:
		if c.RateGroup != nil && *c.RateGroup > 0 {
			return *c.RateGroup
		}
	}
	if c.RateIndividual > 0 {
		return c.RateIndividual
	}
	return c.BaseRate
}

// MaxSchoolPercent возвращает максимальный процент школы из прикрепленных тегов,
// либо значение поля school_percent_tag если тегов нет.
func (c *Client) MaxSchoolPercent() int {
	maxPercent := 0
	for _, t := range c.Tags {
		if t.SchoolPercent > maxPercent {
			maxPercent = t.SchoolPercent
		}
	}
	if maxPercent == 0 && c.SchoolPercentTag > 0 {
		return c.SchoolPercentTag
	}
	return maxPercent
}
