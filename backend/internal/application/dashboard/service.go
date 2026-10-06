package dashboard

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

type ClientRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error)
	ListByTeacherID(ctx context.Context, teacherID uuid.UUID, filter ...crm.ClientFilter) ([]*domain.Client, error)
}

type LessonRepository interface {
	List(ctx context.Context, filter schedule.LessonFilter) ([]*domain.Lesson, error)
}

type ClassroomRepository interface {
	List(ctx context.Context) ([]*domain.Classroom, error)
}

// TodayLessonItem представляет занятие на сегодня в сводке дашборда.
type TodayLessonItem struct {
	ID             uuid.UUID
	ClientID       uuid.UUID
	ClientName     string
	StartAt        time.Time
	EndAt          time.Time
	Format         domain.LessonFormat
	Status         domain.LessonStatus
	LocationType   *string
	OnlineLink     *string
	ClassroomName  *string
	ClassroomColor *string
}

// FinancialSnapshot финансовый срез текущего месяца и оперативные показатели.
type FinancialSnapshot struct {
	MonthEarned        float64
	MonthForecast      float64
	TotalDebts         float64
	ActiveClientsCount int
	WeeklyHours        float64
}

// SummaryResult единая сводка дашборда для главного экрана.
type SummaryResult struct {
	TodayLessons      []*TodayLessonItem
	FinancialSnapshot *FinancialSnapshot
}

type DashboardMetrics struct {
	GrossPotentialRevenue float64 `json:"gross_potential_revenue"`
	NetIncome             float64 `json:"net_income"`
	AverageRate           float64 `json:"average_rate"`
}

type Service struct {
	lessonRepo    LessonRepository
	clientRepo    ClientRepository
	classroomRepo ClassroomRepository
}

func NewService(lessonRepo LessonRepository, clientRepo ClientRepository, classroomRepo ...ClassroomRepository) *Service {
	s := &Service{
		lessonRepo: lessonRepo,
		clientRepo: clientRepo,
	}
	if len(classroomRepo) > 0 {
		s.classroomRepo = classroomRepo[0]
	}
	return s
}

// GetSummary формирует агрегированную сводку для Control Center репетитора.
func (s *Service) GetSummary(ctx context.Context, teacherID uuid.UUID) (*SummaryResult, error) {
	now := time.Now().UTC()

	// 1. Границы сегодняшних суток (00:00:00 - 23:59:59.999 UTC)
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfToday := startOfToday.Add(24 * time.Hour).Add(-time.Nanosecond)

	// 2. Границы текущего календарного месяца
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

	// 3. Границы текущей недели (понедельник 00:00 - воскресенье 23:59)
	weekdayOffset := (int(now.Weekday()) + 6) % 7 // Monday = 0, Sunday = 6
	startOfWeek := startOfToday.AddDate(0, 0, -weekdayOffset)
	endOfWeek := startOfWeek.AddDate(0, 0, 7).Add(-time.Nanosecond)

	// Выгружаем клиентов преподавателя
	clients, err := s.clientRepo.ListByTeacherID(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}
	clientMap := make(map[uuid.UUID]*domain.Client, len(clients))
	for _, c := range clients {
		clientMap[c.ID] = c
	}

	// Выгружаем аудитории (для названий и цветов)
	classroomsMap := make(map[uuid.UUID]*domain.Classroom)
	if s.classroomRepo != nil {
		classrooms, err := s.classroomRepo.List(ctx)
		if err == nil {
			for _, cr := range classrooms {
				classroomsMap[cr.ID] = cr
			}
		}
	}

	// 4. Уроки на сегодня
	todayLessonsRaw, err := s.lessonRepo.List(ctx, schedule.LessonFilter{
		TeacherID: &teacherID,
		From:      &startOfToday,
		To:        &endOfToday,
	})
	if err != nil {
		return nil, fmt.Errorf("list today lessons: %w", err)
	}

	todayLessons := make([]*TodayLessonItem, 0, len(todayLessonsRaw))
	for _, l := range todayLessonsRaw {
		if l.StartTime.Before(startOfToday) || l.StartTime.After(endOfToday) {
			continue
		}

		clientName := "Ученик"
		if c, exists := clientMap[l.ClientID]; exists && c.Name != "" {
			clientName = c.Name
		}

		var locType *string
		var onlineLink *string
		var crName *string
		var crColor *string

		if l.ClassroomID != nil {
			if cr, ok := classroomsMap[*l.ClassroomID]; ok {
				cName := cr.Name
				crName = &cName
				if cr.Color != "" {
					cColor := cr.Color
					crColor = &cColor
				}
			}
			offlineStr := "offline"
			locType = &offlineStr
		} else if l.LocationOrURL != "" {
			onlineStr := "online"
			locType = &onlineStr
			link := l.LocationOrURL
			onlineLink = &link
		}

		todayLessons = append(todayLessons, &TodayLessonItem{
			ID:             l.ID,
			ClientID:       l.ClientID,
			ClientName:     clientName,
			StartAt:        l.StartTime,
			EndAt:          l.EndTime,
			Format:         l.Format,
			Status:         l.Status,
			LocationType:   locType,
			OnlineLink:     onlineLink,
			ClassroomName:  crName,
			ClassroomColor: crColor,
		})
	}

	// 5. Месячные уроки: для расчета month_earned (completed) и month_forecast (будущие scheduled)
	monthLessons, err := s.lessonRepo.List(ctx, schedule.LessonFilter{
		TeacherID: &teacherID,
		From:      &startOfMonth,
		To:        &endOfMonth,
	})
	if err != nil {
		return nil, fmt.Errorf("list month lessons: %w", err)
	}

	var monthEarned float64
	var monthForecast float64

	for _, l := range monthLessons {
		if l.StartTime.Before(startOfMonth) || l.StartTime.After(endOfMonth) {
			continue
		}

		client := clientMap[l.ClientID]
		dur := l.EndTime.Sub(l.StartTime).Seconds() / 3600.0
		rate := 0.0
		if client != nil {
			rate = client.RateForFormat(l.Format)
		}
		revenue := dur * rate

		// Комиссия партнерской школы
		commPercent := 0
		if client != nil {
			for _, tag := range client.Tags {
				if tag.SchoolPercent > commPercent {
					commPercent = tag.SchoolPercent
				}
			}
			if commPercent == 0 && client.SchoolPercentTag > 0 {
				commPercent = client.SchoolPercentTag
			}
		}
		commission := revenue * (float64(commPercent) / 100.0)
		netRevenue := revenue - commission

		if l.Status == domain.StatusCompleted {
			monthEarned += netRevenue
		} else if l.Status == domain.StatusScheduled && (l.StartTime.After(now) || l.StartTime.Equal(now)) {
			// Будущие запланированные уроки до конца текущего месяца
			monthForecast += netRevenue
		}
	}

	// 6. Подсчет дебиторской задолженности клиентов с отрицательным балансом и активных учеников
	var totalDebts float64
	activeClientsMap := make(map[uuid.UUID]bool)

	for _, c := range clients {
		if !c.IsArchived {
			activeClientsMap[c.ID] = true
		}

		// Долги по форматам
		var clientDebt float64
		if c.Balances.IndividualHours < 0 {
			rate := c.RateForFormat(domain.FormatIndividual)
			clientDebt += math.Abs(c.Balances.IndividualHours) * rate
		}
		if c.Balances.PairHours < 0 {
			rate := c.RateForFormat(domain.FormatPair)
			clientDebt += math.Abs(c.Balances.PairHours) * rate
		}
		if c.Balances.GroupHours < 0 {
			rate := c.RateForFormat(domain.FormatGroup)
			clientDebt += math.Abs(c.Balances.GroupHours) * rate
		}
		totalDebts += clientDebt
	}

	// 7. Подсчет запланированных часов на текущую календарную неделю
	weekLessons, err := s.lessonRepo.List(ctx, schedule.LessonFilter{
		TeacherID: &teacherID,
		From:      &startOfWeek,
		To:        &endOfWeek,
	})
	if err != nil {
		return nil, fmt.Errorf("list week lessons: %w", err)
	}

	var weeklyHours float64
	for _, l := range weekLessons {
		if l.StartTime.Before(startOfWeek) || l.StartTime.After(endOfWeek) {
			continue
		}
		if l.Status == domain.StatusScheduled || l.Status == domain.StatusCompleted {
			dur := l.EndTime.Sub(l.StartTime).Seconds() / 3600.0
			weeklyHours += dur
		}
	}

	return &SummaryResult{
		TodayLessons: todayLessons,
		FinancialSnapshot: &FinancialSnapshot{
			MonthEarned:        math.Round(monthEarned*100) / 100,
			MonthForecast:      math.Round(monthForecast*100) / 100,
			TotalDebts:         math.Round(totalDebts*100) / 100,
			ActiveClientsCount: len(activeClientsMap),
			WeeklyHours:        math.Round(weeklyHours*10) / 10,
		},
	}, nil
}

// GetMetrics возвращает финансовые метрики за период (для /dashboard/metrics).
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
