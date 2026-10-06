package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrLessonSeriesNotFound = errors.New("lesson series not found")
	ErrInvalidRRULE         = errors.New("invalid recurrence rule")
	ErrInvalidSeriesDate    = errors.New("until_date must be after or equal to start_date")
	ErrInvalidDuration      = errors.New("duration must be greater than 0")
)

// RecurrenceScope определяет область действия изменений для повторяющихся событий по паттерну Google Calendar.
type RecurrenceScope string

const (
	ScopeThisOnly         RecurrenceScope = "this_only"
	ScopeThisAndFollowing RecurrenceScope = "this_and_following"
	ScopeAllInSeries      RecurrenceScope = "all_in_series"
)

func (s RecurrenceScope) IsValid() bool {
	switch s {
	case ScopeThisOnly, ScopeThisAndFollowing, ScopeAllInSeries:
		return true
	default:
		return false
	}
}

// LessonSeries представляет сущность регулярной серии занятий.
type LessonSeries struct {
	ID              uuid.UUID
	TeacherID       uuid.UUID
	ClientID        uuid.UUID
	ClassroomID     *uuid.UUID
	Title           string
	RRULE           string
	StartTimeOfDay  string // "15:04" или "15:04:00"
	DurationMinutes int
	Format          LessonFormat
	LocationOrURL   string
	Notes           string
	StartDate       time.Time // YYYY-MM-DD
	UntilDate       *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Validate проверяет валидность данных регулярной серии.
func (s *LessonSeries) Validate() error {
	if s.TeacherID == uuid.Nil {
		return errors.New("teacher_id is required")
	}
	if s.ClientID == uuid.Nil {
		return errors.New("client_id is required")
	}
	if strings.TrimSpace(s.Title) == "" {
		return errors.New("title is required")
	}
	if s.DurationMinutes <= 0 {
		return ErrInvalidDuration
	}
	if !s.Format.IsValid() {
		return ErrInvalidLessonFormat
	}
	if strings.TrimSpace(s.RRULE) == "" {
		return ErrInvalidRRULE
	}
	if s.UntilDate != nil && s.UntilDate.Before(s.StartDate) {
		return ErrInvalidSeriesDate
	}
	return nil
}

// RecurrenceRule представляет распарсенное еженедельное правило повторения.
type RecurrenceRule struct {
	Frequency string // WEEKLY
	DaysOfWeek []time.Weekday
}

// ParseWeeklyRRULE парсит строку вида "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR,SA,SU".
func ParseWeeklyRRULE(rrule string) (*RecurrenceRule, error) {
	if !strings.HasPrefix(strings.ToUpper(rrule), "FREQ=WEEKLY") {
		return nil, fmt.Errorf("%w: only WEEKLY frequency is supported", ErrInvalidRRULE)
	}

	parts := strings.Split(rrule, ";")
	var days []time.Weekday

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.ToUpper(strings.TrimSpace(kv[0]))
		val := strings.ToUpper(strings.TrimSpace(kv[1]))

		if key == "BYDAY" {
			dayTokens := strings.Split(val, ",")
			for _, d := range dayTokens {
				switch d {
				case "MO":
					days = append(days, time.Monday)
				case "TU":
					days = append(days, time.Tuesday)
				case "WE":
					days = append(days, time.Wednesday)
				case "TH":
					days = append(days, time.Thursday)
				case "FR":
					days = append(days, time.Friday)
				case "SA":
					days = append(days, time.Saturday)
				case "SU":
					days = append(days, time.Sunday)
				default:
					return nil, fmt.Errorf("%w: unknown day %s", ErrInvalidRRULE, d)
				}
			}
		}
	}

	if len(days) == 0 {
		return nil, fmt.Errorf("%w: no days specified in BYDAY", ErrInvalidRRULE)
	}

	return &RecurrenceRule{
		Frequency:  "WEEKLY",
		DaysOfWeek: days,
	}, nil
}

// BuildWeeklyRRULE собирает строку RRULE по списку дней недели.
func BuildWeeklyRRULE(days []time.Weekday) string {
	dayNames := make([]string, 0, len(days))
	for _, d := range days {
		switch d {
		case time.Monday:
			dayNames = append(dayNames, "MO")
		case time.Tuesday:
			dayNames = append(dayNames, "TU")
		case time.Wednesday:
			dayNames = append(dayNames, "WE")
		case time.Thursday:
			dayNames = append(dayNames, "TH")
		case time.Friday:
			dayNames = append(dayNames, "FR")
		case time.Saturday:
			dayNames = append(dayNames, "SA")
		case time.Sunday:
			dayNames = append(dayNames, "SU")
		}
	}
	return fmt.Sprintf("FREQ=WEEKLY;BYDAY=%s", strings.Join(dayNames, ","))
}

