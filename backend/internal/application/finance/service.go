package finance

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/application/crm"
	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
)

var (
	ErrUnauthorizedAction    = errors.New("unauthorized action")
	ErrPayoutAlreadyExists   = errors.New("partner payout for this period already exists")
	ErrClientNotFound        = domain.ErrClientNotFound
	ErrTagNotFound           = domain.ErrTagNotFound
	ErrPaymentNotFound       = domain.ErrPaymentNotFound
	ErrPartnerPayoutNotFound = domain.ErrPartnerPayoutNotFound
)

// PaymentFilter определяет параметры фильтрации списка платежей.
type PaymentFilter struct {
	TeacherID *uuid.UUID
	ClientID  *uuid.UUID
	From      *time.Time
	To        *time.Time
	Limit     *int
	Offset    *int
}

// PaymentRepository контракт для работы с платежами.
type PaymentRepository interface {
	Create(ctx context.Context, p *domain.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	List(ctx context.Context, filter PaymentFilter) ([]*domain.Payment, error)
	SumAmountByPeriod(ctx context.Context, teacherID uuid.UUID, from, to time.Time) (float64, error)
}

// PartnerPayoutRepository контракт для работы с выплатами партнерам.
type PartnerPayoutRepository interface {
	Create(ctx context.Context, p *domain.PartnerPayout) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PartnerPayout, error)
	GetByTagAndPeriod(ctx context.Context, teacherID, tagID uuid.UUID, periodMonth string) (*domain.PartnerPayout, error)
	ListByPeriod(ctx context.Context, teacherID uuid.UUID, periodMonth string) ([]*domain.PartnerPayout, error)
	ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.PartnerPayout, error)
}

// ClientRepository контракт для работы с клиентами.
type ClientRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error)
	ListByTeacherID(ctx context.Context, teacherID uuid.UUID, filter ...crm.ClientFilter) ([]*domain.Client, error)
}

// SubscriptionRepository контракт для работы с абонементами.
type SubscriptionRepository interface {
	GetByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.ClientSubscription, error)
	Create(ctx context.Context, sub *domain.ClientSubscription) error
	Update(ctx context.Context, sub *domain.ClientSubscription) error
}

// LessonRepository контракт для работы с расписанием.
type LessonRepository interface {
	List(ctx context.Context, filter schedule.LessonFilter) ([]*domain.Lesson, error)
}

// TagRepository контракт для работы с тегами партнерских школ.
type TagRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Tag, error)
	ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.Tag, error)
}

// Transactor интерфейс для выполнения транзакций базы данных.
type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}

// Service предоставляет бизнес-логику финансового учета, взаиморасчетов со школами и экспорта.
type Service struct {
	paymentRepo PaymentRepository
	payoutRepo  PartnerPayoutRepository
	clientRepo  ClientRepository
	subRepo     SubscriptionRepository
	lessonRepo  LessonRepository
	tagRepo     TagRepository
	transactor  Transactor
}

// NewService создает новый экземпляр FinanceService.
func NewService(
	paymentRepo PaymentRepository,
	payoutRepo PartnerPayoutRepository,
	clientRepo ClientRepository,
	subRepo SubscriptionRepository,
	lessonRepo LessonRepository,
	tagRepo TagRepository,
	transactor ...Transactor,
) *Service {
	s := &Service{
		paymentRepo: paymentRepo,
		payoutRepo:  payoutRepo,
		clientRepo:  clientRepo,
		subRepo:     subRepo,
		lessonRepo:  lessonRepo,
		tagRepo:     tagRepo,
	}
	if len(transactor) > 0 {
		s.transactor = transactor[0]
	}
	return s
}

// FinanceSummary сводные финансовые показатели за месяц.
type FinanceSummary struct {
	Month                    string
	TotalPayments            float64
	TotalEarned              float64
	TotalDebts               float64
	TotalCommissions         float64
	ActiveSubscriptionsCount int
	DebtorsCount             int
}

// parseMonthRange парсит месяц формата YYYY-MM в интервал [start, end].
func parseMonthRange(monthStr string) (string, time.Time, time.Time, error) {
	month := strings.TrimSpace(monthStr)
	if month == "" {
		month = time.Now().UTC().Format("2006-01")
	}

	if !domain.ValidatePeriodMonth(month) {
		return "", time.Time{}, time.Time{}, domain.ErrInvalidPeriodMonth
	}

	start, err := time.Parse("2006-01", month)
	if err != nil {
		return "", time.Time{}, time.Time{}, fmt.Errorf("parse month: %w", err)
	}
	start = start.UTC()
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

	return month, start, end, nil
}

// GetFinanceSummary рассчитывает финансовую сводку репетитора за указанный месяц.
func (s *Service) GetFinanceSummary(ctx context.Context, teacherID uuid.UUID, monthStr string) (*FinanceSummary, error) {
	month, startOfMonth, endOfMonth, err := parseMonthRange(monthStr)
	if err != nil {
		return nil, err
	}

	// 1. Поступило оплат в рублях за месяц
	totalPayments, err := s.paymentRepo.SumAmountByPeriod(ctx, teacherID, startOfMonth, endOfMonth)
	if err != nil {
		return nil, fmt.Errorf("sum payments: %w", err)
	}

	// 2. Все клиенты преподавателя
	clients, err := s.clientRepo.ListByTeacherID(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}

	clientMap := make(map[uuid.UUID]*domain.Client, len(clients))
	var totalDebts float64
	var debtorsCount int
	var activeSubsCount int

	for _, c := range clients {
		clientMap[c.ID] = c

		// Расчет долгов по каждому формату (если баланс < 0)
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

		if clientDebt > 0 {
			debtorsCount++
			totalDebts += clientDebt
		}

		// Активный абонемент — клиент с положительным суммарным балансом часов
		if c.Balances.TotalHours > 0 {
			activeSubsCount++
		}
	}

	// 3. Проведенные уроки за месяц для расчета отработанного дохода и партнерских комиссий
	statusCompleted := domain.StatusCompleted
	lessons, err := s.lessonRepo.List(ctx, schedule.LessonFilter{
		TeacherID: &teacherID,
		Status:    &statusCompleted,
		From:      &startOfMonth,
		To:        &endOfMonth,
	})
	if err != nil {
		return nil, fmt.Errorf("list completed lessons: %w", err)
	}

	var totalEarned float64
	var totalCommissions float64

	for _, lesson := range lessons {
		client := clientMap[lesson.ClientID]
		if client == nil {
			continue
		}

		durationHours := lesson.EndTime.Sub(lesson.StartTime).Seconds() / 3600.0
		rate := client.RateForFormat(lesson.Format)
		revenue := durationHours * rate
		totalEarned += revenue

		// Комиссия партнерским школам
		for _, tag := range client.Tags {
			if tag.SchoolPercent > 0 {
				commission := revenue * (float64(tag.SchoolPercent) / 100.0)
				totalCommissions += commission
			}
		}
	}

	return &FinanceSummary{
		Month:                    month,
		TotalPayments:            totalPayments,
		TotalEarned:              totalEarned,
		TotalDebts:               totalDebts,
		TotalCommissions:         totalCommissions,
		ActiveSubscriptionsCount: activeSubsCount,
		DebtorsCount:             debtorsCount,
	}, nil
}
