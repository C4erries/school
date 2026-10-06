package analytics

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/application/crm"
	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
)

// LessonRepository определяет контракт доступа к расписанию уроков.
type LessonRepository interface {
	List(ctx context.Context, filter schedule.LessonFilter) ([]*domain.Lesson, error)
}

// ClientRepository определяет контракт доступа к клиентам преподавателя.
type ClientRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error)
	ListByTeacherID(ctx context.Context, teacherID uuid.UUID, filter ...crm.ClientFilter) ([]*domain.Client, error)
}

// Service предоставляет бизнес-логику аналитики и статистики репетитора.
type Service struct {
	lessonRepo LessonRepository
	clientRepo ClientRepository
}

// NewService создает новый экземпляр AnalyticsService.
func NewService(lessonRepo LessonRepository, clientRepo ClientRepository) *Service {
	return &Service{
		lessonRepo: lessonRepo,
		clientRepo: clientRepo,
	}
}

// OverviewResult сводные KPI за период.
type OverviewResult struct {
	From                time.Time
	To                  time.Time
	TotalLessons        int
	CompletedLessons    int
	CancelledLessons    int
	CompletionRate      float32
	CompletedHours      float32
	GrossRevenue        float32
	NetIncome           float32
	EffectiveHourlyRate float32
}

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

// resolveDateRange определяет границы интервала по умолчанию (текущий календарный месяц).
func resolveDateRange(from, to *time.Time) (time.Time, time.Time) {
	now := time.Now().UTC()
	var startTime, endTime time.Time

	if from != nil && !from.IsZero() {
		startTime = from.UTC()
	} else {
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	}

	if to != nil && !to.IsZero() {
		endTime = to.UTC()
	} else {
		// Конец месяца
		startOfNextMonth := time.Date(startTime.Year(), startTime.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
		endTime = startOfNextMonth.Add(-time.Nanosecond)
	}

	if endTime.Before(startTime) {
		endTime = startTime
	}

	return startTime, endTime
}

func calculateLessonRevenueAndCommission(lesson *domain.Lesson, client *domain.Client) (revenue, commission float64, durationHours float64) {
	dur := lesson.EndTime.Sub(lesson.StartTime).Seconds() / 3600.0
	if dur <= 0 {
		return 0, 0, 0
	}
	durationHours = dur

	rate := 0.0
	if client != nil {
		rate = client.RateForFormat(lesson.Format)
	}
	revenue = dur * rate

	if client != nil {
		for _, tag := range client.Tags {
			if tag.SchoolPercent > 0 {
				commission += revenue * (float64(tag.SchoolPercent) / 100.0)
			}
		}
	}

	return revenue, commission, durationHours
}

// GetOverview вычисляет ключевые показатели эффективности за период.
func (s *Service) GetOverview(ctx context.Context, teacherID uuid.UUID, from, to *time.Time) (*OverviewResult, error) {
	startRange, endRange := resolveDateRange(from, to)

	lessons, err := s.lessonRepo.List(ctx, schedule.LessonFilter{
		TeacherID: &teacherID,
		From:      &startRange,
		To:        &endRange,
	})
	if err != nil {
		return nil, fmt.Errorf("list lessons for overview: %w", err)
	}

	clients, err := s.clientRepo.ListByTeacherID(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list clients for overview: %w", err)
	}
	clientMap := make(map[uuid.UUID]*domain.Client, len(clients))
	for _, c := range clients {
		clientMap[c.ID] = c
	}

	var totalLessons, completedLessons, cancelledLessons int
	var completedHours, grossRevenue, netIncome float64

	for _, l := range lessons {
		if l.StartTime.Before(startRange) || l.StartTime.After(endRange) {
			continue
		}
		totalLessons++

		switch l.Status {
		case domain.StatusCancelled:
			cancelledLessons++
		case domain.StatusCompleted:
			completedLessons++
			client := clientMap[l.ClientID]
			rev, comm, dur := calculateLessonRevenueAndCommission(l, client)
			completedHours += dur
			grossRevenue += rev
			netIncome += (rev - comm)
		}
	}

	completionRate := float32(0)
	if (completedLessons + cancelledLessons) > 0 {
		completionRate = float32(completedLessons) / float32(completedLessons+cancelledLessons) * 100.0
	}

	effectiveHourlyRate := float32(0)
	if completedHours > 0 {
		effectiveHourlyRate = float32(netIncome / completedHours)
	}

	return &OverviewResult{
		From:                startRange,
		To:                  endRange,
		TotalLessons:        totalLessons,
		CompletedLessons:    completedLessons,
		CancelledLessons:    cancelledLessons,
		CompletionRate:      roundFloat32(completionRate, 2),
		CompletedHours:      roundFloat32(float32(completedHours), 2),
		GrossRevenue:        roundFloat32(float32(grossRevenue), 2),
		NetIncome:           roundFloat32(float32(netIncome), 2),
		EffectiveHourlyRate: roundFloat32(effectiveHourlyRate, 2),
	}, nil
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
		// Нормализуем к понедельнику недели startRange
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
		// interval == "month"
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
		hours float64
		income float64
		count int
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
	// Инициализируем для всех клиентов учителя, чтобы они присутствовали
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

func roundFloat32(val float32, precision int) float32 {
	ratio := math.Pow(10, float64(precision))
	return float32(math.Round(float64(val)*ratio) / ratio)
}
