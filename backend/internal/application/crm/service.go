package crm

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
)

type ClientRepository interface {
	Create(ctx context.Context, client *domain.Client) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error)
	ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.Client, error)
	Update(ctx context.Context, client *domain.Client) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *domain.ClientSubscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ClientSubscription, error)
	GetByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.ClientSubscription, error)
	Update(ctx context.Context, sub *domain.ClientSubscription) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	clientRepo ClientRepository
	subRepo    SubscriptionRepository
}

func NewService(clientRepo ClientRepository, subRepo SubscriptionRepository) *Service {
	return &Service{
		clientRepo: clientRepo,
		subRepo:    subRepo,
	}
}

func (s *Service) CreateClient(ctx context.Context, teacherID uuid.UUID, name string, phone *string, baseRate float64, schoolPercentTag int) (*domain.Client, error) {
	client := &domain.Client{
		ID:               uuid.New(),
		TeacherID:        teacherID,
		Name:             name,
		Phone:            phone,
		BaseRate:         baseRate,
		SchoolPercentTag: schoolPercentTag,
		CreatedAt:        time.Now().UTC(),
	}

	if err := s.clientRepo.Create(ctx, client); err != nil {
		return nil, err
	}
	return client, nil
}

func (s *Service) ListClients(ctx context.Context, teacherID uuid.UUID) ([]*domain.Client, error) {
	return s.clientRepo.ListByTeacherID(ctx, teacherID)
}

func (s *Service) GetClient(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	return s.clientRepo.GetByID(ctx, id)
}

func (s *Service) CreateSubscription(ctx context.Context, clientID uuid.UUID, subType domain.SubscriptionType, balance float64) (*domain.ClientSubscription, error) {
	sub := &domain.ClientSubscription{
		ID:        uuid.New(),
		ClientID:  clientID,
		Type:      subType,
		Balance:   balance,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *Service) ListSubscriptions(ctx context.Context, clientID uuid.UUID) ([]*domain.ClientSubscription, error) {
	return s.subRepo.GetByClientID(ctx, clientID)
}
