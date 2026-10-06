package schedule

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
)

type CreateSeriesInput struct {
	TeacherID       uuid.UUID
	ClientID        uuid.UUID
	ClassroomID     *uuid.UUID
	Title           string
	RRULE           string
	StartTimeOfDay  string
	DurationMinutes int
	Format          domain.LessonFormat
	LocationOrURL   string
	Notes           string
	StartDate       time.Time
	UntilDate       *time.Time
}

type UpdateSeriesInput struct {
	SeriesID        uuid.UUID
	CallerID        uuid.UUID
	CallerRole      domain.Role
	ClientID        *uuid.UUID
	ClassroomID     *uuid.UUID
	ClearClassroom  bool
	Title           *string
	RRULE           *string
	StartTimeOfDay  *string
	DurationMinutes *int
	Format          *domain.LessonFormat
	LocationOrURL   *string
	Notes           *string
	UntilDate       *time.Time
	ClearUntilDate  bool
}

// CreateSeries создает новую регулярную серию занятий.
func (s *Service) CreateSeries(ctx context.Context, input CreateSeriesInput) (*domain.LessonSeries, error) {
	client, err := s.clientRepo.GetByID(ctx, input.ClientID)
	if err != nil {
		return nil, fmt.Errorf("get client: %w", err)
	}
	if client.TeacherID != input.TeacherID {
		return nil, domain.ErrUnauthorizedLessonAction
	}

	if input.ClassroomID != nil && *input.ClassroomID != uuid.Nil {
		if _, err := s.classroomRepo.GetByID(ctx, *input.ClassroomID); err != nil {
			return nil, err
		}
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "Занятие"
	}

	now := time.Now().UTC()
	series := &domain.LessonSeries{
		ID:              uuid.New(),
		TeacherID:       input.TeacherID,
		ClientID:        input.ClientID,
		ClassroomID:     input.ClassroomID,
		Title:           title,
		RRULE:           strings.TrimSpace(input.RRULE),
		StartTimeOfDay:  strings.TrimSpace(input.StartTimeOfDay),
		DurationMinutes: input.DurationMinutes,
		Format:          input.Format,
		LocationOrURL:   strings.TrimSpace(input.LocationOrURL),
		Notes:           strings.TrimSpace(input.Notes),
		StartDate:       input.StartDate,
		UntilDate:       input.UntilDate,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := series.Validate(); err != nil {
		return nil, err
	}

	if s.seriesRepo == nil {
		return nil, fmt.Errorf("series repository is not configured")
	}

	if err := s.seriesRepo.Create(ctx, series); err != nil {
		return nil, fmt.Errorf("create lesson series: %w", err)
	}

	return series, nil
}

// GetSeries возвращает серию по ID.
func (s *Service) GetSeries(ctx context.Context, id, callerID uuid.UUID, callerRole domain.Role) (*domain.LessonSeries, error) {
	if s.seriesRepo == nil {
		return nil, fmt.Errorf("series repository is not configured")
	}

	series, err := s.seriesRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if callerRole != domain.RoleOwner && series.TeacherID != callerID {
		return nil, domain.ErrUnauthorizedLessonAction
	}

	return series, nil
}

// ListSeries возвращает все серии преподавателя.
func (s *Service) ListSeries(ctx context.Context, teacherID uuid.UUID) ([]*domain.LessonSeries, error) {
	if s.seriesRepo == nil {
		return nil, fmt.Errorf("series repository is not configured")
	}

	return s.seriesRepo.ListByTeacherID(ctx, teacherID)
}

// UpdateSeries обновляет глобальные параметры серии (scope = all_in_series).
func (s *Service) UpdateSeries(ctx context.Context, input UpdateSeriesInput) (*domain.LessonSeries, error) {
	if s.seriesRepo == nil {
		return nil, fmt.Errorf("series repository is not configured")
	}

	series, err := s.seriesRepo.GetByID(ctx, input.SeriesID)
	if err != nil {
		return nil, err
	}

	if input.CallerRole != domain.RoleOwner && series.TeacherID != input.CallerID {
		return nil, domain.ErrUnauthorizedLessonAction
	}

	if input.ClientID != nil {
		client, err := s.clientRepo.GetByID(ctx, *input.ClientID)
		if err != nil {
			return nil, fmt.Errorf("get client: %w", err)
		}
		if client.TeacherID != series.TeacherID {
			return nil, domain.ErrUnauthorizedLessonAction
		}
		series.ClientID = *input.ClientID
	}

	if input.ClearClassroom {
		series.ClassroomID = nil
	} else if input.ClassroomID != nil {
		if *input.ClassroomID == uuid.Nil {
			series.ClassroomID = nil
		} else {
			if _, err := s.classroomRepo.GetByID(ctx, *input.ClassroomID); err != nil {
				return nil, err
			}
			series.ClassroomID = input.ClassroomID
		}
	}

	if input.Title != nil {
		t := strings.TrimSpace(*input.Title)
		if t != "" {
			series.Title = t
		}
	}

	if input.RRULE != nil {
		series.RRULE = strings.TrimSpace(*input.RRULE)
	}
	if input.StartTimeOfDay != nil {
		series.StartTimeOfDay = strings.TrimSpace(*input.StartTimeOfDay)
	}
	if input.DurationMinutes != nil {
		series.DurationMinutes = *input.DurationMinutes
	}
	if input.Format != nil {
		series.Format = *input.Format
	}
	if input.LocationOrURL != nil {
		series.LocationOrURL = strings.TrimSpace(*input.LocationOrURL)
	}
	if input.Notes != nil {
		series.Notes = strings.TrimSpace(*input.Notes)
	}
	if input.ClearUntilDate {
		series.UntilDate = nil
	} else if input.UntilDate != nil {
		series.UntilDate = input.UntilDate
	}

	if err := series.Validate(); err != nil {
		return nil, err
	}

	series.UpdatedAt = time.Now().UTC()
	if err := s.seriesRepo.Update(ctx, series); err != nil {
		return nil, fmt.Errorf("update lesson series: %w", err)
	}

	return series, nil
}

// DeleteSeries удаляет серию (все будущие уроки).
func (s *Service) DeleteSeries(ctx context.Context, id, callerID uuid.UUID, callerRole domain.Role) error {
	if s.seriesRepo == nil {
		return fmt.Errorf("series repository is not configured")
	}

	series, err := s.seriesRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if callerRole != domain.RoleOwner && series.TeacherID != callerID {
		return domain.ErrUnauthorizedLessonAction
	}

	return s.seriesRepo.Delete(ctx, id)
}

