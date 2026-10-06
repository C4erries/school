package schedule

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
)

type CreateClassroomInput struct {
	Name        string
	Capacity    int
	Color       string
	Description string
}

type UpdateClassroomInput struct {
	ID          uuid.UUID
	Name        string
	Capacity    int
	Color       string
	Description string
}

// --- Кабинеты (Classrooms) ---

func (s *Service) CreateClassroom(ctx context.Context, input CreateClassroomInput) (*domain.Classroom, error) {
	name, err := domain.ValidateClassroomInput(input.Name, input.Capacity)
	if err != nil {
		return nil, err
	}

	color := strings.TrimSpace(input.Color)
	if color == "" {
		color = "#3B82F6"
	}

	classroom := &domain.Classroom{
		ID:          uuid.New(),
		Name:        name,
		Capacity:    input.Capacity,
		Color:       color,
		Description: strings.TrimSpace(input.Description),
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.classroomRepo.Create(ctx, classroom); err != nil {
		return nil, fmt.Errorf("create classroom: %w", err)
	}

	return classroom, nil
}

func (s *Service) GetClassroom(ctx context.Context, id uuid.UUID) (*domain.Classroom, error) {
	return s.classroomRepo.GetByID(ctx, id)
}

func (s *Service) ListClassrooms(ctx context.Context) ([]*domain.Classroom, error) {
	return s.classroomRepo.List(ctx)
}

func (s *Service) UpdateClassroom(ctx context.Context, input UpdateClassroomInput) (*domain.Classroom, error) {
	name, err := domain.ValidateClassroomInput(input.Name, input.Capacity)
	if err != nil {
		return nil, err
	}

	existing, err := s.classroomRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	color := strings.TrimSpace(input.Color)
	if color == "" {
		color = existing.Color
	}

	existing.Name = name
	existing.Capacity = input.Capacity
	existing.Color = color
	existing.Description = strings.TrimSpace(input.Description)

	if err := s.classroomRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("update classroom: %w", err)
	}

	return existing, nil
}

func (s *Service) DeleteClassroom(ctx context.Context, id uuid.UUID) error {
	return s.classroomRepo.Delete(ctx, id)
}

