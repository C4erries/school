package dashboard

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
)

type ClientRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error)
}

type LessonRepository interface {
	List(ctx context.Context, filter schedule.LessonFilter) ([]*domain.Lesson, error)
}

type DashboardMetrics struct {
	GrossPotentialRevenue float64 `json:"gross_potential_revenue"`
	NetIncome             float64 `json:"net_income"`
	AverageRate           float64 `json:"average_rate"`
}

type Service struct {
	lessonRepo LessonRepository
	clientRepo ClientRepository
}

func NewService(lessonRepo LessonRepository, clientRepo ClientRepository) *Service {
	return &Service{
		lessonRepo: lessonRepo,
		clientRepo: clientRepo,
	}
}

func (s *Service) GetMetrics(ctx context.Context, teacherID uuid.UUID, from, to time.Time) (*DashboardMetrics, error) {
	lessons, err := s.lessonRepo.List(ctx, schedule.LessonFilter{
		TeacherID: &teacherID,
		From:      &from,
		To:        &to,
	})
	if err != nil {
		return nil, err
	}

	var gross float64
	var net float64
	var totalHours float64
	var totalRate float64
	var uniqueClients = make(map[uuid.UUID]bool)

	for _, lesson := range lessons {
		if lesson.Status == domain.StatusCancelled {
			continue
		}

		client, err := s.clientRepo.GetByID(ctx, lesson.ClientID)
		if err != nil {
			continue
		}

		durationHours := lesson.EndTime.Sub(lesson.StartTime).Seconds() / 3600.0
		clientRate := client.RateForFormat(lesson.Format)
		lessonRevenue := durationHours * clientRate

		gross += lessonRevenue

		schoolPercent := client.MaxSchoolPercent()
		commission := lessonRevenue * (float64(schoolPercent) / 100.0)
		net += (lessonRevenue - commission)

		rate := client.RateIndividual
		if rate == 0 {
			rate = client.BaseRate
		}
		if !uniqueClients[client.ID] {
			uniqueClients[client.ID] = true
			totalRate += rate
		}
		totalHours += durationHours
	}

	var avgRate float64
	if len(uniqueClients) > 0 {
		avgRate = totalRate / float64(len(uniqueClients))
	}

	return &DashboardMetrics{
		GrossPotentialRevenue: gross,
		NetIncome:             net,
		AverageRate:           avgRate,
	}, nil
}
