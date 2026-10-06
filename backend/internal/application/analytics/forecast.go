package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
)

// FormatForecast прогноз нагрузки и дохода по формату.
type FormatForecast struct {
	Format  domain.LessonFormat
	Hours   float32
	Revenue float32
}

// ForecastResult прогнозные показатели расписания за период.
type ForecastResult struct {
	From                      time.Time
	To                        time.Time
	ScheduledLessons          int
	ScheduledHours            float32
	GrossPotentialRevenue     float32
	PartnerCommissionExpected float32
	NetPotentialIncome        float32
	ByFormat                  []FormatForecast
}

// GetForecast рассчитывает прогнозную нагрузку и доход по урокам в статусе scheduled.
func (s *Service) GetForecast(ctx context.Context, teacherID uuid.UUID, from, to *time.Time) (*ForecastResult, error) {
	startRange, endRange := resolveDateRange(from, to)

	statusScheduled := domain.StatusScheduled
	lessons, err := s.lessonRepo.List(ctx, schedule.LessonFilter{
		TeacherID: &teacherID,
		Status:    &statusScheduled,
		From:      &startRange,
		To:        &endRange,
	})
	if err != nil {
		return nil, fmt.Errorf("list scheduled lessons for forecast: %w", err)
	}

	clients, err := s.clientRepo.ListByTeacherID(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list clients for forecast: %w", err)
	}
	clientMap := make(map[uuid.UUID]*domain.Client, len(clients))
	for _, c := range clients {
		clientMap[c.ID] = c
	}

	var scheduledLessons int
	var scheduledHours float64
	var grossPotentialRevenue float64
	var partnerCommissionExpected float64

	type formatAcc struct {
		hours   float64
		revenue float64
	}
	byFormatMap := map[domain.LessonFormat]*formatAcc{
		domain.FormatIndividual: {},
		domain.FormatPair:       {},
		domain.FormatGroup:      {},
	}

	for _, l := range lessons {
		if l.StartTime.Before(startRange) || l.StartTime.After(endRange) {
			continue
		}
		scheduledLessons++

		client := clientMap[l.ClientID]
		rev, comm, dur := calculateLessonRevenueAndCommission(l, client)

		scheduledHours += dur
		grossPotentialRevenue += rev
		partnerCommissionExpected += comm

		acc, ok := byFormatMap[l.Format]
		if !ok {
			acc = &formatAcc{}
			byFormatMap[l.Format] = acc
		}
		acc.hours += dur
		acc.revenue += rev
	}

	netPotentialIncome := grossPotentialRevenue - partnerCommissionExpected

	formatOrder := []domain.LessonFormat{
		domain.FormatIndividual,
		domain.FormatPair,
		domain.FormatGroup,
	}

	byFormat := make([]FormatForecast, 0, len(formatOrder))
	for _, f := range formatOrder {
		acc := byFormatMap[f]
		byFormat = append(byFormat, FormatForecast{
			Format:  f,
			Hours:   roundFloat32(float32(acc.hours), 2),
			Revenue: roundFloat32(float32(acc.revenue), 2),
		})
	}

	return &ForecastResult{
		From:                      startRange,
		To:                        endRange,
		ScheduledLessons:          scheduledLessons,
		ScheduledHours:            roundFloat32(float32(scheduledHours), 2),
		GrossPotentialRevenue:     roundFloat32(float32(grossPotentialRevenue), 2),
		PartnerCommissionExpected: roundFloat32(float32(partnerCommissionExpected), 2),
		NetPotentialIncome:        roundFloat32(float32(netPotentialIncome), 2),
		ByFormat:                  byFormat,
	}, nil
}

