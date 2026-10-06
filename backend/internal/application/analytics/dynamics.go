package analytics

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
)

// DynamicsPoint точка временного ряда динамики нагрузки и доходов.
type DynamicsPoint struct {
	Label          string
	From           time.Time
	To             time.Time
	CompletedHours float32
	NetIncome      float32
	CompletedCount int
	CancelledCount int
}

// FormatStat статистика по формату занятий.
type FormatStat struct {
	Format              domain.LessonFormat
	CompletedHours      float32
	NetIncome           float32
	LessonsCount        int
	HoursSharePercent   float32
	RevenueSharePercent float32
}

// ClientStat показатели ученика за период.
type ClientStat struct {
	ClientID       uuid.UUID
	ClientName     string
	CompletedHours float32
	NetIncome      float32
	CompletedCount int
	CancelledCount int
	AttendanceRate float32
}

// monthRu возвращает краткое название месяца на русском языке.
func monthRu(m time.Month) string {
	months := [...]string{
		"янв", "фев", "мар", "апр", "май", "июн",
		"июл", "авг", "сен", "окт", "ноя", "дек",
	}
	if m >= 1 && m <= 12 {
		return months[m-1]
	}
	return ""
}

// monthNameRu возвращает полное название месяца на русском языке.
func monthNameRu(m time.Month) string {
	months := [...]string{
		"Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
		"Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь",
	}
	if m >= 1 && m <= 12 {
		return months[m-1]
	}
	return ""
}

// GetDynamics строит временной ряд динамики нагрузки и доходов (шаг: week или month).
func (s *Service) GetDynamics(ctx context.Context, teacherID uuid.UUID, interval string, from, to *time.Time) ([]DynamicsPoint, error) {
	startRange, endRange := resolveDateRange(from, to)

	lessons, err := s.lessonRepo.List(ctx, schedule.LessonFilter{
		TeacherID: &teacherID,
		From:      &startRange,
		To:        &endRange,
	})
	if err != nil {
		return nil, fmt.Errorf("list lessons for dynamics: %w", err)
	}

	clients, err := s.clientRepo.ListByTeacherID(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list clients for dynamics: %w", err)
	}
	clientMap := make(map[uuid.UUID]*domain.Client, len(clients))
	for _, c := range clients {
		clientMap[c.ID] = c
	}

	interval = strings.ToLower(strings.TrimSpace(interval))
	if interval != "week" && interval != "month" {
		interval = "month"
	}

	type timeBucket struct {
		from  time.Time
		to    time.Time
		label string
	}

	var buckets []timeBucket

	if interval == "week" {
		weekday := startRange.Weekday()
		daysFromMonday := int(weekday - time.Monday)
		if daysFromMonday < 0 {
			daysFromMonday = 6 // Sunday
		}
		curStart := time.Date(startRange.Year(), startRange.Month(), startRange.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -daysFromMonday)

		for curStart.Before(endRange) || curStart.Equal(endRange) {
			curEnd := curStart.AddDate(0, 0, 7).Add(-time.Nanosecond)
			_, weekNum := curStart.ISOWeek()
			label := fmt.Sprintf("Неделя %d (%d %s - %d %s)",
				weekNum,
				curStart.Day(), monthRu(curStart.Month()),
				curEnd.Day(), monthRu(curEnd.Month()),
			)
			if curStart.Month() == curEnd.Month() {
				label = fmt.Sprintf("Неделя %d (%d - %d %s)",
					weekNum,
					curStart.Day(),
					curEnd.Day(),
					monthRu(curStart.Month()),
				)
			}

			buckets = append(buckets, timeBucket{
				from:  curStart,
				to:    curEnd,
				label: label,
			})

			curStart = curStart.AddDate(0, 0, 7)
		}
	} else {
		curStart := time.Date(startRange.Year(), startRange.Month(), 1, 0, 0, 0, 0, time.UTC)
		for curStart.Before(endRange) || curStart.Equal(endRange) {
			curEnd := curStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
			label := fmt.Sprintf("%s %d", monthNameRu(curStart.Month()), curStart.Year())

			buckets = append(buckets, timeBucket{
				from:  curStart,
				to:    curEnd,
				label: label,
			})

			curStart = curStart.AddDate(0, 1, 0)
		}
	}

	result := make([]DynamicsPoint, len(buckets))
	for i, b := range buckets {
		var compCount, cancCount int
		var compHours, netInc float64

		for _, l := range lessons {
			if (l.StartTime.Equal(b.from) || l.StartTime.After(b.from)) &&
				(l.StartTime.Equal(b.to) || l.StartTime.Before(b.to)) {
				if l.Status == domain.StatusCancelled {
					cancCount++
				} else if l.Status == domain.StatusCompleted {
					compCount++
					client := clientMap[l.ClientID]
					rev, comm, dur := calculateLessonRevenueAndCommission(l, client)
					compHours += dur
					netInc += (rev - comm)
				}
			}
		}

		result[i] = DynamicsPoint{
			Label:          b.label,
			From:           b.from,
			To:             b.to,
			CompletedHours: roundFloat32(float32(compHours), 2),
			NetIncome:      roundFloat32(float32(netInc), 2),
			CompletedCount: compCount,
			CancelledCount: cancCount,
		}
	}

	return result, nil
}

// GetFormats рассчитывает распределение часов и выручки по форматам занятий.
func (s *Service) GetFormats(ctx context.Context, teacherID uuid.UUID, from, to *time.Time) ([]FormatStat, error) {
	startRange, endRange := resolveDateRange(from, to)

	lessons, err := s.lessonRepo.List(ctx, schedule.LessonFilter{
		TeacherID: &teacherID,
		From:      &startRange,
		To:        &endRange,
	})
	if err != nil {
		return nil, fmt.Errorf("list lessons for formats: %w", err)
	}

	clients, err := s.clientRepo.ListByTeacherID(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list clients for formats: %w", err)
	}
	clientMap := make(map[uuid.UUID]*domain.Client, len(clients))
	for _, c := range clients {
		clientMap[c.ID] = c
	}

	type formatAccum struct {
		hours  float64
		income float64
		count  int
	}

	accum := map[domain.LessonFormat]*formatAccum{
		domain.FormatIndividual: {},
		domain.FormatPair:       {},
		domain.FormatGroup:      {},
	}

	var totalHours, totalIncome float64

	for _, l := range lessons {
		if l.StartTime.Before(startRange) || l.StartTime.After(endRange) {
			continue
		}
		if l.Status != domain.StatusCompleted {
			continue
		}

		acc, ok := accum[l.Format]
		if !ok {
			acc = &formatAccum{}
			accum[l.Format] = acc
		}

		client := clientMap[l.ClientID]
		rev, comm, dur := calculateLessonRevenueAndCommission(l, client)
		net := rev - comm

		acc.hours += dur
		acc.income += net
		acc.count++

		totalHours += dur
		totalIncome += net
	}

	order := []domain.LessonFormat{
		domain.FormatIndividual,
		domain.FormatPair,
		domain.FormatGroup,
	}

	result := make([]FormatStat, 0, len(order))
	for _, fmtKey := range order {
		acc := accum[fmtKey]
		hoursShare := float32(0)
		if totalHours > 0 {
			hoursShare = float32(acc.hours / totalHours * 100.0)
		}
		revenueShare := float32(0)
		if totalIncome > 0 {
			revenueShare = float32(acc.income / totalIncome * 100.0)
		}

		result = append(result, FormatStat{
			Format:              fmtKey,
			CompletedHours:      roundFloat32(float32(acc.hours), 2),
			NetIncome:           roundFloat32(float32(acc.income), 2),
			LessonsCount:        acc.count,
			HoursSharePercent:   roundFloat32(hoursShare, 1),
			RevenueSharePercent: roundFloat32(revenueShare, 1),
		})
	}

	return result, nil
}

// GetClients вычисляет рейтинг и показатели учеников за период.
func (s *Service) GetClients(ctx context.Context, teacherID uuid.UUID, sortField string, limit int, from, to *time.Time) ([]ClientStat, error) {
	startRange, endRange := resolveDateRange(from, to)

	lessons, err := s.lessonRepo.List(ctx, schedule.LessonFilter{
		TeacherID: &teacherID,
		From:      &startRange,
		To:        &endRange,
	})
	if err != nil {
		return nil, fmt.Errorf("list lessons for client stats: %w", err)
	}

	clients, err := s.clientRepo.ListByTeacherID(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list clients for client stats: %w", err)
	}
	clientMap := make(map[uuid.UUID]*domain.Client, len(clients))
	for _, c := range clients {
		clientMap[c.ID] = c
	}

	type clientAccum struct {
		completedHours float64
		netIncome      float64
		completedCount int
		cancelledCount int
	}

	statsMap := make(map[uuid.UUID]*clientAccum, len(clients))
	for _, c := range clients {
		statsMap[c.ID] = &clientAccum{}
	}

	for _, l := range lessons {
		if l.StartTime.Before(startRange) || l.StartTime.After(endRange) {
			continue
		}
		acc, ok := statsMap[l.ClientID]
		if !ok {
			acc = &clientAccum{}
			statsMap[l.ClientID] = acc
		}

		if l.Status == domain.StatusCancelled {
			acc.cancelledCount++
		} else if l.Status == domain.StatusCompleted {
			acc.completedCount++
			client := clientMap[l.ClientID]
			rev, comm, dur := calculateLessonRevenueAndCommission(l, client)
			acc.completedHours += dur
			acc.netIncome += (rev - comm)
		}
	}

	var stats []ClientStat
	for clientID, acc := range statsMap {
		client := clientMap[clientID]
		name := "Ученик"
		if client != nil {
			name = client.Name
		}

		attRate := float32(0)
		total := acc.completedCount + acc.cancelledCount
		if total > 0 {
			attRate = float32(acc.completedCount) / float32(total) * 100.0
		}

		stats = append(stats, ClientStat{
			ClientID:       clientID,
			ClientName:     name,
			CompletedHours: roundFloat32(float32(acc.completedHours), 2),
			NetIncome:      roundFloat32(float32(acc.netIncome), 2),
			CompletedCount: acc.completedCount,
			CancelledCount: acc.cancelledCount,
			AttendanceRate: roundFloat32(attRate, 1),
		})
	}

	sortField = strings.ToLower(strings.TrimSpace(sortField))
	switch sortField {
	case "revenue":
		sort.Slice(stats, func(i, j int) bool {
			if stats[i].NetIncome == stats[j].NetIncome {
				return stats[i].CompletedHours > stats[j].CompletedHours
			}
			return stats[i].NetIncome > stats[j].NetIncome
		})
	case "cancellations":
		sort.Slice(stats, func(i, j int) bool {
			if stats[i].CancelledCount == stats[j].CancelledCount {
				return stats[i].CompletedHours > stats[j].CompletedHours
			}
			return stats[i].CancelledCount > stats[j].CancelledCount
		})
	default: // "hours"
		sort.Slice(stats, func(i, j int) bool {
			if stats[i].CompletedHours == stats[j].CompletedHours {
				return stats[i].NetIncome > stats[j].NetIncome
			}
			return stats[i].CompletedHours > stats[j].CompletedHours
		})
	}

	if limit <= 0 {
		limit = 20
	}
	if len(stats) > limit {
		stats = stats[:limit]
	}

	return stats, nil
}

