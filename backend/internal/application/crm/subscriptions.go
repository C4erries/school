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

