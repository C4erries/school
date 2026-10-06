package analytics

import (
	"context"
	"fmt"
	"math"
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

func roundFloat32(val float32, precision int) float32 {
	ratio := math.Pow(10, float64(precision))
	return float32(math.Round(float64(val)*ratio) / ratio)
}
