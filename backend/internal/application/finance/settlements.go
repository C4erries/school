package finance

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

