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

var (
	ErrUnauthorizedAction = errors.New("unauthorized action")
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

type TagRepository interface {
	Create(ctx context.Context, tag *domain.Tag) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Tag, error)
	ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.Tag, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AssignToClient(ctx context.Context, clientID, tagID uuid.UUID) error
	RemoveFromClient(ctx context.Context, clientID, tagID uuid.UUID) error
	SetClientTags(ctx context.Context, clientID uuid.UUID, tagIDs []uuid.UUID) error
}

type Service struct {
	clientRepo ClientRepository
	subRepo    SubscriptionRepository
	tagRepo    TagRepository
}

func NewService(clientRepo ClientRepository, subRepo SubscriptionRepository, tagRepo ...TagRepository) *Service {
	s := &Service{
		clientRepo: clientRepo,
		subRepo:    subRepo,
	}
	if len(tagRepo) > 0 {
		s.tagRepo = tagRepo[0]
	}
	return s
}

type CreateClientInput struct {
	TeacherID        uuid.UUID
	Name             string
	Phone            *string
	BaseRate         float64
	RateIndividual   float64
	RatePair         *float64
	RateGroup        *float64
	SchoolPercentTag int
	TagIDs           []uuid.UUID
}

type UpdateClientInput struct {
	ID             uuid.UUID
	CallerID       uuid.UUID
	CallerRole     domain.Role
	Name           *string
	Phone          *string
	RateIndividual *float64
	RatePair       *float64
	RateGroup      *float64
	TagIDs         *[]uuid.UUID
}

func (s *Service) CreateClient(ctx context.Context, teacherID uuid.UUID, name string, phone *string, baseRate float64, schoolPercentTag int) (*domain.Client, error) {
	client := &domain.Client{
		ID:               uuid.New(),
		TeacherID:        teacherID,
		Name:             strings.TrimSpace(name),
		Phone:            phone,
		BaseRate:         baseRate,
		RateIndividual:   baseRate,
		SchoolPercentTag: schoolPercentTag,
		Tags:             make([]domain.Tag, 0),
		Balances:         domain.ClientBalances{},
		CreatedAt:        time.Now().UTC(),
	}

	if err := s.clientRepo.Create(ctx, client); err != nil {
		return nil, err
	}
	return client, nil
}

func (s *Service) CreateClientWithRates(ctx context.Context, input CreateClientInput) (*domain.Client, error) {
	rateIndiv := input.RateIndividual
	if rateIndiv == 0 && input.BaseRate > 0 {
		rateIndiv = input.BaseRate
	}

	client := &domain.Client{
		ID:               uuid.New(),
		TeacherID:        input.TeacherID,
		Name:             strings.TrimSpace(input.Name),
		Phone:            input.Phone,
		BaseRate:         rateIndiv,
		RateIndividual:   rateIndiv,
		RatePair:         input.RatePair,
		RateGroup:        input.RateGroup,
		SchoolPercentTag: input.SchoolPercentTag,
		Tags:             make([]domain.Tag, 0),
		Balances:         domain.ClientBalances{},
		CreatedAt:        time.Now().UTC(),
	}

	if err := s.clientRepo.Create(ctx, client); err != nil {
		return nil, err
	}

	if len(input.TagIDs) > 0 && s.tagRepo != nil {
		if err := s.tagRepo.SetClientTags(ctx, client.ID, input.TagIDs); err != nil {
			return nil, fmt.Errorf("set client tags: %w", err)
		}
		return s.clientRepo.GetByID(ctx, client.ID)
	}

	return client, nil
}

func (s *Service) UpdateClient(ctx context.Context, input UpdateClientInput) (*domain.Client, error) {
	client, err := s.clientRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.CallerRole != domain.RoleOwner && client.TeacherID != input.CallerID {
		return nil, ErrUnauthorizedAction
	}

	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed != "" {
			client.Name = trimmed
		}
	}

	if input.Phone != nil {
		client.Phone = input.Phone
	}

	if input.RateIndividual != nil {
		client.RateIndividual = *input.RateIndividual
		client.BaseRate = *input.RateIndividual
	}

	if input.RatePair != nil {
		client.RatePair = input.RatePair
	}

	if input.RateGroup != nil {
		client.RateGroup = input.RateGroup
	}

	if err := s.clientRepo.Update(ctx, client); err != nil {
		return nil, fmt.Errorf("update client: %w", err)
	}

	if input.TagIDs != nil && s.tagRepo != nil {
		if err := s.tagRepo.SetClientTags(ctx, client.ID, *input.TagIDs); err != nil {
			return nil, fmt.Errorf("set client tags: %w", err)
		}
	}

	return s.clientRepo.GetByID(ctx, client.ID)
}

func (s *Service) DeleteClient(ctx context.Context, id, callerID uuid.UUID, callerRole domain.Role) error {
	client, err := s.clientRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if callerRole != domain.RoleOwner && client.TeacherID != callerID {
		return ErrUnauthorizedAction
	}

	return s.clientRepo.Delete(ctx, id)
}

func (s *Service) ListClients(ctx context.Context, teacherID uuid.UUID) ([]*domain.Client, error) {
	return s.clientRepo.ListByTeacherID(ctx, teacherID)
}

func (s *Service) GetClient(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	return s.clientRepo.GetByID(ctx, id)
}

func (s *Service) CreateSubscription(ctx context.Context, clientID uuid.UUID, format domain.SubscriptionFormat, balance float64) (*domain.ClientSubscription, error) {
	sub := &domain.ClientSubscription{
		ID:        uuid.New(),
		ClientID:  clientID,
		Format:    format,
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

// --- Методы для работы с тегами ---

func (s *Service) CreateTag(ctx context.Context, teacherID uuid.UUID, name string, schoolPercent int, color string) (*domain.Tag, error) {
	if s.tagRepo == nil {
		return nil, errors.New("tag repo not initialized")
	}

	validName, err := domain.ValidateTagInput(name, schoolPercent)
	if err != nil {
		return nil, err
	}

	color = strings.TrimSpace(color)
	if color == "" {
		color = "indigo"
	}

	tag := &domain.Tag{
		ID:            uuid.New(),
		TeacherID:     teacherID,
		Name:          validName,
		SchoolPercent: schoolPercent,
		Color:         color,
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.tagRepo.Create(ctx, tag); err != nil {
		return nil, fmt.Errorf("create tag: %w", err)
	}

	return tag, nil
}

func (s *Service) ListTags(ctx context.Context, teacherID uuid.UUID) ([]*domain.Tag, error) {
	if s.tagRepo == nil {
		return nil, errors.New("tag repo not initialized")
	}
	return s.tagRepo.ListByTeacherID(ctx, teacherID)
}

func (s *Service) DeleteTag(ctx context.Context, tagID, callerID uuid.UUID, callerRole domain.Role) error {
	if s.tagRepo == nil {
		return errors.New("tag repo not initialized")
	}

	tag, err := s.tagRepo.GetByID(ctx, tagID)
	if err != nil {
		return err
	}

	if callerRole != domain.RoleOwner && tag.TeacherID != callerID {
		return domain.ErrUnauthorizedTagAction
	}

	return s.tagRepo.Delete(ctx, tagID)
}

func (s *Service) AssignTagToClient(ctx context.Context, clientID, tagID, callerID uuid.UUID, callerRole domain.Role) (*domain.Client, error) {
	if s.tagRepo == nil {
		return nil, errors.New("tag repo not initialized")
	}

	client, err := s.clientRepo.GetByID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	if callerRole != domain.RoleOwner && client.TeacherID != callerID {
		return nil, ErrUnauthorizedAction
	}

	tag, err := s.tagRepo.GetByID(ctx, tagID)
	if err != nil {
		return nil, err
	}

	if callerRole != domain.RoleOwner && tag.TeacherID != callerID {
		return nil, domain.ErrUnauthorizedTagAction
	}

	if err := s.tagRepo.AssignToClient(ctx, clientID, tagID); err != nil {
		return nil, fmt.Errorf("assign tag to client: %w", err)
	}

	return s.clientRepo.GetByID(ctx, clientID)
}

func (s *Service) RemoveTagFromClient(ctx context.Context, clientID, tagID, callerID uuid.UUID, callerRole domain.Role) (*domain.Client, error) {
	if s.tagRepo == nil {
		return nil, errors.New("tag repo not initialized")
	}

	client, err := s.clientRepo.GetByID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	if callerRole != domain.RoleOwner && client.TeacherID != callerID {
		return nil, ErrUnauthorizedAction
	}

	if err := s.tagRepo.RemoveFromClient(ctx, clientID, tagID); err != nil {
		return nil, fmt.Errorf("remove tag from client: %w", err)
	}

	return s.clientRepo.GetByID(ctx, clientID)
}
