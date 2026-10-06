package finance

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/application/crm"
	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
)

var (
	ErrUnauthorizedAction   = errors.New("unauthorized action")
	ErrPayoutAlreadyExists  = errors.New("partner payout for this period already exists")
	ErrClientNotFound       = domain.ErrClientNotFound
	ErrTagNotFound          = domain.ErrTagNotFound
	ErrPaymentNotFound      = domain.ErrPaymentNotFound
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

// CreatePaymentInput входные параметры для регистрации платежа.
type CreatePaymentInput struct {
	TeacherID     uuid.UUID
	ClientID      uuid.UUID
	Amount        float64
	Hours         float64
	Format        domain.SubscriptionFormat
	PaymentMethod domain.PaymentMethod
	PaidAt        *time.Time
	Notes         string
	CallerRole    domain.Role
}

// PaymentItem элемент журнала оплат с именем ученика.
type PaymentItem struct {
	Payment    *domain.Payment
	ClientName string
}

// FinanceSummary сводные финансовые показатели за месяц.
type FinanceSummary struct {
	Month                   string
	TotalPayments           float64
	TotalEarned             float64
	TotalDebts              float64
	TotalCommissions        float64
	ActiveSubscriptionsCount int
	DebtorsCount            int
}

// PartnerSettlementItem результат взаиморасчета с партнерской школой по тегу за месяц.
type PartnerSettlementItem struct {
	TagID            uuid.UUID
	TagName          string
	TagColor         string
	SchoolPercent    int
	PeriodMonth      string
	LessonsCount     int
	GrossAmount      float64
	CommissionAmount float64
	IsPaid           bool
	PaidAt           *time.Time
	PayoutID         *uuid.UUID
}

// CreatePartnerPayoutInput входные параметры для фиксации выплаты партнеру.
type CreatePartnerPayoutInput struct {
	TeacherID        uuid.UUID
	TagID            uuid.UUID
	PeriodMonth      string
	GrossAmount      float64
	CommissionAmount float64
	PaidAt           *time.Time
	Notes            string
	CallerRole       domain.Role
}

// CreatePayment регистрирует факт оплаты ученика и автоматически пополняет баланс абонемента нужного формата.
func (s *Service) CreatePayment(ctx context.Context, input CreatePaymentInput) (*PaymentItem, error) {
	client, err := s.clientRepo.GetByID(ctx, input.ClientID)
	if err != nil {
		return nil, fmt.Errorf("get client: %w", err)
	}

	if input.CallerRole != domain.RoleOwner && client.TeacherID != input.TeacherID {
		return nil, ErrUnauthorizedAction
	}

	paidAt := time.Now().UTC()
	if input.PaidAt != nil && !input.PaidAt.IsZero() {
		paidAt = input.PaidAt.UTC()
	}

	method := input.PaymentMethod
	if method == "" {
		method = domain.PaymentMethodTransfer
	}

	payment := &domain.Payment{
		ID:            uuid.New(),
		TeacherID:     client.TeacherID,
		ClientID:      input.ClientID,
		Amount:        input.Amount,
		Hours:         input.Hours,
		Format:        input.Format,
		PaymentMethod: method,
		PaidAt:        paidAt,
		Notes:         strings.TrimSpace(input.Notes),
		CreatedAt:     time.Now().UTC(),
	}

	if err := payment.Validate(); err != nil {
		return nil, err
	}

	execFn := func(txCtx context.Context) error {
		if err := s.paymentRepo.Create(txCtx, payment); err != nil {
			return fmt.Errorf("create payment record: %w", err)
		}

		if s.subRepo != nil {
			subs, err := s.subRepo.GetByClientID(txCtx, input.ClientID)
			if err != nil {
				return fmt.Errorf("get subscriptions: %w", err)
			}

			var matchingSub *domain.ClientSubscription
			for _, sub := range subs {
				if sub.Format == input.Format {
					matchingSub = sub
					break
				}
			}

			if matchingSub != nil {
				matchingSub.Balance += input.Hours
				if err := s.subRepo.Update(txCtx, matchingSub); err != nil {
					return fmt.Errorf("update subscription balance: %w", err)
				}
			} else {
				newSub := &domain.ClientSubscription{
					ID:        uuid.New(),
					ClientID:  input.ClientID,
					Format:    input.Format,
					Balance:   input.Hours,
					CreatedAt: time.Now().UTC(),
				}
				if err := s.subRepo.Create(txCtx, newSub); err != nil {
					return fmt.Errorf("create subscription: %w", err)
				}
			}
		}

		return nil
	}

	if s.transactor != nil {
		if err := s.transactor.WithinTransaction(ctx, execFn); err != nil {
			return nil, err
		}
	} else {
		if err := execFn(ctx); err != nil {
			return nil, err
		}
	}

	return &PaymentItem{
		Payment:    payment,
		ClientName: client.Name,
	}, nil
}

// ListPayments возвращает список платежей с возможностью фильтрации и пагинации.
func (s *Service) ListPayments(ctx context.Context, filter PaymentFilter) ([]*PaymentItem, error) {
	payments, err := s.paymentRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}

	var teacherID uuid.UUID
	if filter.TeacherID != nil {
		teacherID = *filter.TeacherID
	}

	clientNames := make(map[uuid.UUID]string)
	if teacherID != uuid.Nil && s.clientRepo != nil {
		clients, err := s.clientRepo.ListByTeacherID(ctx, teacherID)
		if err == nil {
			for _, c := range clients {
				clientNames[c.ID] = c.Name
			}
		}
	}

	result := make([]*PaymentItem, 0, len(payments))
	for _, p := range payments {
		name := clientNames[p.ClientID]
		if name == "" && s.clientRepo != nil {
			if c, err := s.clientRepo.GetByID(ctx, p.ClientID); err == nil && c != nil {
				name = c.Name
				clientNames[p.ClientID] = name
			}
		}
		result = append(result, &PaymentItem{
			Payment:    p,
			ClientName: name,
		})
	}

	return result, nil
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
		Month:                   month,
		TotalPayments:           totalPayments,
		TotalEarned:             totalEarned,
		TotalDebts:              totalDebts,
		TotalCommissions:        totalCommissions,
		ActiveSubscriptionsCount: activeSubsCount,
		DebtorsCount:            debtorsCount,
	}, nil
}

// ListPartnerSettlements рассчитывает взаиморасчеты с партнерскими школами по тегам за месяц.
func (s *Service) ListPartnerSettlements(ctx context.Context, teacherID uuid.UUID, monthStr string) ([]*PartnerSettlementItem, error) {
	month, startOfMonth, endOfMonth, err := parseMonthRange(monthStr)
	if err != nil {
		return nil, err
	}

	// Теги репетитора
	tags, err := s.tagRepo.ListByTeacherID(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}

	// Выплаты за данный месяц
	payouts, err := s.payoutRepo.ListByPeriod(ctx, teacherID, month)
	if err != nil {
		return nil, fmt.Errorf("list partner payouts: %w", err)
	}

	payoutMap := make(map[uuid.UUID]*domain.PartnerPayout, len(payouts))
	for _, p := range payouts {
		payoutMap[p.TagID] = p
	}

	// Проведенные уроки за месяц
	statusCompleted := domain.StatusCompleted
	lessons, err := s.lessonRepo.List(ctx, schedule.LessonFilter{
		TeacherID: &teacherID,
		Status:    &statusCompleted,
		From:      &startOfMonth,
		To:        &endOfMonth,
	})
	if err != nil {
		return nil, fmt.Errorf("list lessons: %w", err)
	}

	// Клиенты репетитора
	clients, err := s.clientRepo.ListByTeacherID(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}
	clientMap := make(map[uuid.UUID]*domain.Client, len(clients))
	for _, c := range clients {
		clientMap[c.ID] = c
	}

	// Статистика по тегам
	type tagStat struct {
		lessonsCount     int
		grossAmount      float64
		commissionAmount float64
	}
	stats := make(map[uuid.UUID]*tagStat)
	for _, t := range tags {
		stats[t.ID] = &tagStat{}
	}

	for _, lesson := range lessons {
		client := clientMap[lesson.ClientID]
		if client == nil {
			continue
		}

		durationHours := lesson.EndTime.Sub(lesson.StartTime).Seconds() / 3600.0
		rate := client.RateForFormat(lesson.Format)
		revenue := durationHours * rate

		for _, tag := range client.Tags {
			stat, ok := stats[tag.ID]
			if !ok {
				stat = &tagStat{}
				stats[tag.ID] = stat
			}
			stat.lessonsCount++
			stat.grossAmount += revenue
			commission := revenue * (float64(tag.SchoolPercent) / 100.0)
			stat.commissionAmount += commission
		}
	}

	items := make([]*PartnerSettlementItem, 0, len(tags))
	for _, tag := range tags {
		if tag.SchoolPercent <= 0 {
			continue
		}

		stat := stats[tag.ID]
		if stat == nil {
			stat = &tagStat{}
		}

		item := &PartnerSettlementItem{
			TagID:            tag.ID,
			TagName:          tag.Name,
			TagColor:         tag.Color,
			SchoolPercent:    tag.SchoolPercent,
			PeriodMonth:      month,
			LessonsCount:     stat.lessonsCount,
			GrossAmount:      stat.grossAmount,
			CommissionAmount: stat.commissionAmount,
			IsPaid:           false,
		}

		if payout, exists := payoutMap[tag.ID]; exists {
			item.IsPaid = true
			paidAt := payout.PaidAt
			item.PaidAt = &paidAt
			payoutID := payout.ID
			item.PayoutID = &payoutID
		}

		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].TagName < items[j].TagName
	})

	return items, nil
}

// CreatePartnerPayout фиксирует факт выплаты партнерской школе за месяц.
func (s *Service) CreatePartnerPayout(ctx context.Context, input CreatePartnerPayoutInput) (*domain.PartnerPayout, error) {
	tag, err := s.tagRepo.GetByID(ctx, input.TagID)
	if err != nil {
		return nil, fmt.Errorf("get tag: %w", err)
	}

	if input.CallerRole != domain.RoleOwner && tag.TeacherID != input.TeacherID {
		return nil, ErrUnauthorizedAction
	}

	if !domain.ValidatePeriodMonth(input.PeriodMonth) {
		return nil, domain.ErrInvalidPeriodMonth
	}

	// Проверяем, не была ли уже зафиксирована выплата за этот период
	existing, err := s.payoutRepo.GetByTagAndPeriod(ctx, tag.TeacherID, input.TagID, input.PeriodMonth)
	if err == nil && existing != nil {
		return nil, ErrPayoutAlreadyExists
	}

	paidAt := time.Now().UTC()
	if input.PaidAt != nil && !input.PaidAt.IsZero() {
		paidAt = input.PaidAt.UTC()
	}

	payout := &domain.PartnerPayout{
		ID:               uuid.New(),
		TeacherID:        tag.TeacherID,
		TagID:            input.TagID,
		PeriodMonth:      input.PeriodMonth,
		GrossAmount:      input.GrossAmount,
		CommissionAmount: input.CommissionAmount,
		PaidAt:           paidAt,
		Notes:            strings.TrimSpace(input.Notes),
		CreatedAt:        time.Now().UTC(),
	}

	if err := payout.Validate(); err != nil {
		return nil, err
	}

	if err := s.payoutRepo.Create(ctx, payout); err != nil {
		return nil, fmt.Errorf("create partner payout: %w", err)
	}

	return payout, nil
}

// --- CSV Экспорт данных (UTF-8 BOM) ---

const utf8BOM = "\xEF\xBB\xBF"

// ExportClientsCSV формирует CSV файл реестра учеников с кодировкой UTF-8 BOM.
func (s *Service) ExportClientsCSV(ctx context.Context, teacherID uuid.UUID) ([]byte, error) {
	clients, err := s.clientRepo.ListByTeacherID(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString(utf8BOM)
	w := csv.NewWriter(&buf)

	header := []string{
		"ID",
		"Имя",
		"Телефон",
		"Базовая ставка (₽)",
		"Ставка индивид. (₽)",
		"Ставка пара (₽)",
		"Ставка группа (₽)",
		"Баланс индивид. (ч)",
		"Баланс пара (ч)",
		"Баланс группа (ч)",
		"Общий баланс (ч)",
		"Теги",
		"В архиве",
		"Дата создания",
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	for _, c := range clients {
		phone := ""
		if c.Phone != nil {
			phone = *c.Phone
		}
		ratePair := ""
		if c.RatePair != nil {
			ratePair = fmt.Sprintf("%.2f", *c.RatePair)
		}
		rateGroup := ""
		if c.RateGroup != nil {
			rateGroup = fmt.Sprintf("%.2f", *c.RateGroup)
		}

		tagNames := make([]string, 0, len(c.Tags))
		for _, t := range c.Tags {
			tagNames = append(tagNames, t.Name)
		}

		isArchived := "Нет"
		if c.IsArchived {
			isArchived = "Да"
		}

		row := []string{
			c.ID.String(),
			c.Name,
			phone,
			fmt.Sprintf("%.2f", c.BaseRate),
			fmt.Sprintf("%.2f", c.RateIndividual),
			ratePair,
			rateGroup,
			strconv.FormatFloat(c.Balances.IndividualHours, 'f', 2, 64),
			strconv.FormatFloat(c.Balances.PairHours, 'f', 2, 64),
			strconv.FormatFloat(c.Balances.GroupHours, 'f', 2, 64),
			strconv.FormatFloat(c.Balances.TotalHours, 'f', 2, 64),
			strings.Join(tagNames, ", "),
			isArchived,
			c.CreatedAt.Format("2006-01-02 15:04"),
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ExportLessonsCSV формирует CSV файл расписания уроков с фильтром по датам и UTF-8 BOM.
func (s *Service) ExportLessonsCSV(ctx context.Context, teacherID uuid.UUID, from, to *time.Time) ([]byte, error) {
	lessons, err := s.lessonRepo.List(ctx, schedule.LessonFilter{
		TeacherID: &teacherID,
		From:      from,
		To:        to,
	})
	if err != nil {
		return nil, fmt.Errorf("list lessons: %w", err)
	}

	clientNames := make(map[uuid.UUID]string)
	if s.clientRepo != nil {
		if clients, err := s.clientRepo.ListByTeacherID(ctx, teacherID); err == nil {
			for _, c := range clients {
				clientNames[c.ID] = c.Name
			}
		}
	}

	var buf bytes.Buffer
	buf.WriteString(utf8BOM)
	w := csv.NewWriter(&buf)

	header := []string{
		"ID",
		"Дата начала",
		"Дата окончания",
		"Тема",
		"Формат",
		"Ученик",
		"Статус",
		"Локация / Ссылка",
		"Заметки",
		"Причина отмены",
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	for _, l := range lessons {
		clientName := clientNames[l.ClientID]
		if clientName == "" && s.clientRepo != nil {
			if c, err := s.clientRepo.GetByID(ctx, l.ClientID); err == nil && c != nil {
				clientName = c.Name
				clientNames[l.ClientID] = clientName
			}
		}

		formatName := string(l.Format)
		switch l.Format {
		case domain.FormatIndividual:
			formatName = "Индивидуальное"
		case domain.FormatPair:
			formatName = "Парное"
		case domain.FormatGroup:
			formatName = "Мини-группа"
		}

		statusName := string(l.Status)
		switch l.Status {
		case domain.StatusScheduled:
			statusName = "Запланирован"
		case domain.StatusCompleted:
			statusName = "Завершен"
		case domain.StatusCancelled:
			statusName = "Отменен"
		}

		row := []string{
			l.ID.String(),
			l.StartTime.Format("2006-01-02 15:04"),
			l.EndTime.Format("2006-01-02 15:04"),
			l.Title,
			formatName,
			clientName,
			statusName,
			l.LocationOrURL,
			l.Notes,
			l.CancelReason,
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ExportPaymentsCSV формирует CSV файл журнала оплат с фильтром по датам и UTF-8 BOM.
func (s *Service) ExportPaymentsCSV(ctx context.Context, teacherID uuid.UUID, from, to *time.Time) ([]byte, error) {
	payments, err := s.paymentRepo.List(ctx, PaymentFilter{
		TeacherID: &teacherID,
		From:      from,
		To:        to,
	})
	if err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}

	clientNames := make(map[uuid.UUID]string)
	if s.clientRepo != nil {
		if clients, err := s.clientRepo.ListByTeacherID(ctx, teacherID); err == nil {
			for _, c := range clients {
				clientNames[c.ID] = c.Name
			}
		}
	}

	var buf bytes.Buffer
	buf.WriteString(utf8BOM)
	w := csv.NewWriter(&buf)

	header := []string{
		"ID",
		"Дата оплаты",
		"Ученик",
		"Сумма (₽)",
		"Часы",
		"Формат",
		"Способ оплаты",
		"Заметки",
		"Дата создания",
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	for _, p := range payments {
		clientName := clientNames[p.ClientID]
		if clientName == "" && s.clientRepo != nil {
			if c, err := s.clientRepo.GetByID(ctx, p.ClientID); err == nil && c != nil {
				clientName = c.Name
				clientNames[p.ClientID] = clientName
			}
		}

		formatName := string(p.Format)
		switch p.Format {
		case domain.SubscriptionFormatIndividual:
			formatName = "Индивидуальное"
		case domain.SubscriptionFormatPair:
			formatName = "Парное"
		case domain.SubscriptionFormatGroup:
			formatName = "Мини-группа"
		}

		methodName := string(p.PaymentMethod)
		switch p.PaymentMethod {
		case domain.PaymentMethodTransfer:
			methodName = "Перевод / СБП"
		case domain.PaymentMethodCash:
			methodName = "Наличные"
		case domain.PaymentMethodCard:
			methodName = "Банковская карта"
		case domain.PaymentMethodOther:
			methodName = "Другое"
		}

		row := []string{
			p.ID.String(),
			p.PaidAt.Format("2006-01-02 15:04"),
			clientName,
			fmt.Sprintf("%.2f", p.Amount),
			strconv.FormatFloat(p.Hours, 'f', 2, 64),
			formatName,
			methodName,
			p.Notes,
			p.CreatedAt.Format("2006-01-02 15:04"),
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
