package finance

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
)

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

