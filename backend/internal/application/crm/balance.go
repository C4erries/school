package crm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
)

type AdjustBalanceInput struct {
	ClientID   uuid.UUID
	CallerID   uuid.UUID
	CallerRole domain.Role
	Format     domain.SubscriptionFormat
	DeltaHours float64
	Reason     string
}

func (s *Service) AdjustBalance(ctx context.Context, input AdjustBalanceInput) (*domain.Client, error) {
	if !input.Format.IsValid() {
		return nil, domain.ErrInvalidSubscriptionFormat
	}
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		return nil, errors.New("adjustment reason is required")
	}
	if input.DeltaHours == 0 {
		return nil, errors.New("delta hours cannot be zero")
	}

	client, err := s.clientRepo.GetByID(ctx, input.ClientID)
	if err != nil {
		return nil, err
	}

	if input.CallerRole != domain.RoleOwner && client.TeacherID != input.CallerID {
		return nil, ErrUnauthorizedAction
	}

	// 1. Обновляем баланс абонемента соответствующего формата
	if s.subRepo != nil {
		subs, err := s.subRepo.GetByClientID(ctx, input.ClientID)
		if err != nil {
			return nil, fmt.Errorf("get subscriptions: %w", err)
		}

		var matchingSub *domain.ClientSubscription
		for _, sub := range subs {
			if sub.Format == input.Format {
				matchingSub = sub
				break
			}
		}

		if matchingSub != nil {
			matchingSub.Balance += input.DeltaHours
			if err := s.subRepo.Update(ctx, matchingSub); err != nil {
				return nil, fmt.Errorf("update subscription: %w", err)
			}
		} else {
			newSub := &domain.ClientSubscription{
				ID:        uuid.New(),
				ClientID:  input.ClientID,
				Format:    input.Format,
				Balance:   input.DeltaHours,
				CreatedAt: time.Now().UTC(),
			}
			if err := s.subRepo.Create(ctx, newSub); err != nil {
				return nil, fmt.Errorf("create subscription: %w", err)
			}
		}
	}

	// 2. Логируем запись в client_balance_adjustments
	if s.adjRepo != nil {
		adj := &domain.ClientBalanceAdjustment{
			ID:         uuid.New(),
			ClientID:   input.ClientID,
			TeacherID:  client.TeacherID,
			Format:     input.Format,
			DeltaHours: input.DeltaHours,
			Reason:     reason,
			CreatedAt:  time.Now().UTC(),
		}
		if err := s.adjRepo.Create(ctx, adj); err != nil {
			return nil, fmt.Errorf("create balance adjustment: %w", err)
		}
	}

	return s.clientRepo.GetByID(ctx, input.ClientID)
}

