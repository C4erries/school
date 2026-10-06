package schedule

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
)

type UpdateRecurringLessonInput struct {
	LessonID       uuid.UUID
	SeriesID       *uuid.UUID
	OccurrenceDate *time.Time
	Scope          domain.RecurrenceScope
	CallerID       uuid.UUID
	CallerRole     domain.Role
	Title          *string
	ClientID       *uuid.UUID
	ClassroomID    *uuid.UUID
	ClearClassroom bool
	StartTime      *time.Time
	EndTime        *time.Time
	Format         *domain.LessonFormat
	LocationOrURL  *string
	Notes          *string
	CancelReason   *string
}

// UpdateRecurringLesson обновляет урок с учетом scope (Google Calendar Pattern).
func (s *Service) UpdateRecurringLesson(ctx context.Context, input UpdateRecurringLessonInput) (*domain.Lesson, error) {
	scope := input.Scope
	if !scope.IsValid() {
		scope = domain.ScopeThisOnly
	}

	physicalLesson, _ := s.lessonRepo.GetByID(ctx, input.LessonID)
	var effectiveSeriesID *uuid.UUID = input.SeriesID
	var effectiveOccDate *time.Time = input.OccurrenceDate

	if physicalLesson != nil {
		if input.CallerRole != domain.RoleOwner && physicalLesson.TeacherID != input.CallerID {
			return nil, domain.ErrUnauthorizedLessonAction
		}
		if physicalLesson.SeriesID != nil {
			effectiveSeriesID = physicalLesson.SeriesID
		}
		if physicalLesson.OriginalStartTime != nil {
			effectiveOccDate = physicalLesson.OriginalStartTime
		} else {
			effectiveOccDate = &physicalLesson.StartTime
		}
	} else if effectiveSeriesID == nil {
		series, slot, vErr := s.FindVirtualLesson(ctx, input.LessonID, &input.CallerID)
		if vErr == nil && series != nil && slot != nil {
			effectiveSeriesID = &series.ID
			effectiveOccDate = &slot.OriginalStartTime
		}
	}

	// Если это разовый урок без серии
	if effectiveSeriesID == nil {
		if physicalLesson != nil {
			return s.UpdateLesson(ctx, UpdateLessonInput{
				LessonID:       physicalLesson.ID,
				CallerID:       input.CallerID,
				CallerRole:     input.CallerRole,
				Title:          input.Title,
				ClientID:       input.ClientID,
				ClassroomID:    input.ClassroomID,
				ClearClassroom: input.ClearClassroom,
				StartTime:      input.StartTime,
				EndTime:        input.EndTime,
				Format:         input.Format,
				LocationOrURL:  input.LocationOrURL,
				Notes:          input.Notes,
				CancelReason:   input.CancelReason,
			})
		}
		return nil, domain.ErrLessonNotFound
	}

	if s.seriesRepo == nil {
		return nil, fmt.Errorf("series repository is not configured")
	}

	series, err := s.seriesRepo.GetByID(ctx, *effectiveSeriesID)
	if err != nil {
		return nil, err
	}
	if input.CallerRole != domain.RoleOwner && series.TeacherID != input.CallerID {
		return nil, domain.ErrUnauthorizedLessonAction
	}

	switch scope {
	case domain.ScopeAllInSeries:
		// Обновляем параметры всей серии
		var duration *int
		if input.StartTime != nil && input.EndTime != nil {
			dur := int(input.EndTime.Sub(*input.StartTime).Minutes())
			if dur > 0 {
				duration = &dur
			}
		}
		var startTOD *string
		if input.StartTime != nil {
			tod := input.StartTime.Format("15:04")
			startTOD = &tod
		}
		_, err := s.UpdateSeries(ctx, UpdateSeriesInput{
			SeriesID:        series.ID,
			CallerID:        input.CallerID,
			CallerRole:      input.CallerRole,
			ClientID:        input.ClientID,
			ClassroomID:     input.ClassroomID,
			ClearClassroom:  input.ClearClassroom,
			Title:           input.Title,
			StartTimeOfDay:  startTOD,
			DurationMinutes: duration,
			Format:          input.Format,
			LocationOrURL:   input.LocationOrURL,
			Notes:           input.Notes,
		})
		if err != nil {
			return nil, err
		}
		if physicalLesson != nil {
			return s.UpdateLesson(ctx, UpdateLessonInput{
				LessonID:       physicalLesson.ID,
				CallerID:       input.CallerID,
				CallerRole:     input.CallerRole,
				Title:          input.Title,
				ClientID:       input.ClientID,
				ClassroomID:    input.ClassroomID,
				ClearClassroom: input.ClearClassroom,
				StartTime:      input.StartTime,
				EndTime:        input.EndTime,
				Format:         input.Format,
				LocationOrURL:  input.LocationOrURL,
				Notes:          input.Notes,
				CancelReason:   input.CancelReason,
			})
		}
		// Возвращаем виртуальный экземпляр
		slot := OccurrenceSlot{
			StartTime:         series.StartDate,
			EndTime:           series.StartDate.Add(time.Duration(series.DurationMinutes) * time.Minute),
			OriginalStartTime: series.StartDate,
		}
		return BuildVirtualLesson(series, slot), nil

	case domain.ScopeThisAndFollowing:
		// Разделение серии
		if effectiveOccDate == nil && input.StartTime != nil {
			effectiveOccDate = input.StartTime
		}
		if effectiveOccDate == nil {
			return nil, domain.ErrLessonNotFound
		}
		// 1. Обрезаем старую серию
		until := effectiveOccDate.AddDate(0, 0, -1)
		series.UntilDate = &until
		series.UpdatedAt = time.Now().UTC()
		if err := s.seriesRepo.Update(ctx, series); err != nil {
			return nil, fmt.Errorf("split series: %w", err)
		}

		// 2. Создаем новую серию с даты effectiveOccDate
		dur := series.DurationMinutes
		if input.StartTime != nil && input.EndTime != nil {
			dur = int(input.EndTime.Sub(*input.StartTime).Minutes())
		}
		tod := series.StartTimeOfDay
		if input.StartTime != nil {
			tod = input.StartTime.Format("15:04")
		}
		fmtChoice := series.Format
		if input.Format != nil {
			fmtChoice = *input.Format
		}
		classroom := series.ClassroomID
		if input.ClearClassroom {
			classroom = nil
		} else if input.ClassroomID != nil {
			classroom = input.ClassroomID
		}
		locOrURL := series.LocationOrURL
		if input.LocationOrURL != nil {
			locOrURL = *input.LocationOrURL
		}
		title := series.Title
		if input.Title != nil {
			title = *input.Title
		}
		notes := series.Notes
		if input.Notes != nil {
			notes = *input.Notes
		}

		newSeries, err := s.CreateSeries(ctx, CreateSeriesInput{
			TeacherID:       series.TeacherID,
			ClientID:        series.ClientID,
			ClassroomID:     classroom,
			Title:           title,
			RRULE:           series.RRULE,
			StartTimeOfDay:  tod,
			DurationMinutes: dur,
			Format:          fmtChoice,
			LocationOrURL:   locOrURL,
			Notes:           notes,
			StartDate:       *effectiveOccDate,
			UntilDate:       nil,
		})
		if err != nil {
			return nil, fmt.Errorf("create follow series: %w", err)
		}

		slot := OccurrenceSlot{
			StartTime:         *effectiveOccDate,
			EndTime:           effectiveOccDate.Add(time.Duration(dur) * time.Minute),
			OriginalStartTime: *effectiveOccDate,
		}
		return BuildVirtualLesson(newSeries, slot), nil

	case domain.ScopeThisOnly:
		fallthrough
	default:
		// Если урок уже материализован
		if physicalLesson != nil {
			return s.UpdateLesson(ctx, UpdateLessonInput{
				LessonID:       physicalLesson.ID,
				CallerID:       input.CallerID,
				CallerRole:     input.CallerRole,
				Title:          input.Title,
				ClientID:       input.ClientID,
				ClassroomID:    input.ClassroomID,
				ClearClassroom: input.ClearClassroom,
				StartTime:      input.StartTime,
				EndTime:        input.EndTime,
				Format:         input.Format,
				LocationOrURL:  input.LocationOrURL,
				Notes:          input.Notes,
				CancelReason:   input.CancelReason,
			})
		}
		// Материализуем исключение в lessons с series_id и original_start_time
		if effectiveOccDate == nil && input.StartTime != nil {
			effectiveOccDate = input.StartTime
		}
		if effectiveOccDate == nil {
			return nil, domain.ErrLessonNotFound
		}

		startTime := *effectiveOccDate
		if input.StartTime != nil {
			startTime = *input.StartTime
		}
		endTime := startTime.Add(time.Duration(series.DurationMinutes) * time.Minute)
		if input.EndTime != nil {
			endTime = *input.EndTime
		}

		if err := domain.ValidateLessonTimes(startTime, endTime); err != nil {
			return nil, err
		}

		format := series.Format
		if input.Format != nil {
			format = *input.Format
		}

		classroomID := series.ClassroomID
		if input.ClearClassroom {
			classroomID = nil
		} else if input.ClassroomID != nil {
			classroomID = input.ClassroomID
		}

		if classroomID != nil && *classroomID != uuid.Nil {
			collision, err := s.lessonRepo.HasClassroomCollision(
				ctx,
				*classroomID,
				series.TeacherID,
				startTime,
				endTime,
				nil,
			)
			if err != nil {
				return nil, fmt.Errorf("check classroom collision: %w", err)
			}
			if collision {
				return nil, domain.ErrClassroomCollision
			}
		}

		title := series.Title
		if input.Title != nil && strings.TrimSpace(*input.Title) != "" {
			title = strings.TrimSpace(*input.Title)
		}

		locURL := series.LocationOrURL
		if input.LocationOrURL != nil {
			locURL = strings.TrimSpace(*input.LocationOrURL)
		}

		notes := series.Notes
		if input.Notes != nil {
			notes = strings.TrimSpace(*input.Notes)
		}

		now := time.Now().UTC()
		mat := &domain.Lesson{
			ID:                uuid.New(),
			TeacherID:         series.TeacherID,
			ClientID:          series.ClientID,
			Title:             title,
			ClassroomID:       classroomID,
			StartTime:         startTime,
			EndTime:           endTime,
			Format:            format,
			LocationOrURL:     locURL,
			Status:            domain.StatusScheduled,
			Notes:             notes,
			SeriesID:          &series.ID,
			OriginalStartTime: effectiveOccDate,
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		if err := s.lessonRepo.Create(ctx, mat); err != nil {
			return nil, fmt.Errorf("materialize updated lesson: %w", err)
		}

		return mat, nil
	}
}

