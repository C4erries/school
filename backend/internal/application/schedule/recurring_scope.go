package schedule

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
)

// CompleteRecurringLesson материализует виртуальный урок или завершает физический со списанием баланса абонемента.
func (s *Service) CompleteRecurringLesson(
	ctx context.Context,
	lessonID uuid.UUID,
	seriesID *uuid.UUID,
	occurrenceDate *time.Time,
	callerID uuid.UUID,
	callerRole domain.Role,
) (*domain.Lesson, error) {
	// 1. Проверяем, существует ли физический урок в БД
	lesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err == nil {
		// Физический урок уже существует — штатно завершаем
		return s.CompleteLesson(ctx, lesson.ID, callerID, callerRole)
	}

	// 2. Если урок виртуальный (не найден в БД), проверяем наличие seriesID и occurrenceDate
	if seriesID == nil || occurrenceDate == nil || s.seriesRepo == nil {
		return nil, domain.ErrLessonNotFound
	}

	series, err := s.seriesRepo.GetByID(ctx, *seriesID)
	if err != nil {
		return nil, err
	}

	if callerRole != domain.RoleOwner && series.TeacherID != callerID {
		return nil, domain.ErrUnauthorizedLessonAction
	}

	// Вычисляем слот
	hour, minute, err := ParseTimeOfDay(series.StartTimeOfDay)
	if err != nil {
		return nil, err
	}

	loc := occurrenceDate.Location()
	startTime := time.Date(occurrenceDate.Year(), occurrenceDate.Month(), occurrenceDate.Day(), hour, minute, 0, 0, loc)
	endTime := startTime.Add(time.Duration(series.DurationMinutes) * time.Minute)

	now := time.Now().UTC()
	materializedLesson := &domain.Lesson{
		ID:                uuid.New(),
		TeacherID:         series.TeacherID,
		ClientID:          series.ClientID,
		Title:             series.Title,
		ClassroomID:       series.ClassroomID,
		StartTime:         startTime,
		EndTime:           endTime,
		Format:            series.Format,
		LocationOrURL:     series.LocationOrURL,
		Status:            domain.StatusCompleted,
		Notes:             series.Notes,
		SeriesID:          &series.ID,
		OriginalStartTime: &startTime,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.lessonRepo.Create(ctx, materializedLesson); err != nil {
		return nil, fmt.Errorf("materialize completed lesson: %w", err)
	}

	// Списываем часы с абонемента ученика
	if s.subRepo != nil {
		durationHours := float64(series.DurationMinutes) / 60.0
		targetFormat := domain.SubscriptionFormat(series.Format)

		subs, err := s.subRepo.GetByClientID(ctx, series.ClientID)
		if err == nil {
			var matchingSub *domain.ClientSubscription
			for _, sub := range subs {
				if sub.Format == targetFormat {
					matchingSub = sub
					break
				}
			}

			if matchingSub != nil {
				matchingSub.Balance -= durationHours
				_ = s.subRepo.Update(ctx, matchingSub)
			} else {
				newSub := &domain.ClientSubscription{
					ID:        uuid.New(),
					ClientID:  series.ClientID,
					Format:    targetFormat,
					Balance:   -durationHours,
					CreatedAt: time.Now().UTC(),
				}
				_ = s.subRepo.Create(ctx, newSub)
			}
		}
	}

	return materializedLesson, nil
}

// CancelRecurringLesson обрабатывает отмену с учетом scope (Google Calendar Pattern).
func (s *Service) CancelRecurringLesson(
	ctx context.Context,
	lessonID uuid.UUID,
	seriesID *uuid.UUID,
	occurrenceDate *time.Time,
	scope domain.RecurrenceScope,
	reason string,
	callerID uuid.UUID,
	callerRole domain.Role,
) error {
	// По умолчанию если scope не передан, то this_only
	if !scope.IsValid() {
		scope = domain.ScopeThisOnly
	}

	// Ищем физический урок, если есть
	physicalLesson, _ := s.lessonRepo.GetByID(ctx, lessonID)

	var effectiveSeriesID *uuid.UUID = seriesID
	var effectiveOccDate *time.Time = occurrenceDate

	if physicalLesson != nil {
		if callerRole != domain.RoleOwner && physicalLesson.TeacherID != callerID {
			return domain.ErrUnauthorizedLessonAction
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
		series, slot, vErr := s.FindVirtualLesson(ctx, lessonID, &callerID)
		if vErr == nil && series != nil && slot != nil {
			effectiveSeriesID = &series.ID
			effectiveOccDate = &slot.OriginalStartTime
		}
	}

	// Если это разовый урок без серии или серия не указана
	if effectiveSeriesID == nil {
		if physicalLesson != nil {
			_, err := s.CancelLesson(ctx, physicalLesson.ID, callerID, callerRole, reason)
			return err
		}
		return domain.ErrLessonNotFound
	}

	if s.seriesRepo == nil {
		return fmt.Errorf("series repository is not configured")
	}

	series, err := s.seriesRepo.GetByID(ctx, *effectiveSeriesID)
	if err != nil {
		return err
	}
	if callerRole != domain.RoleOwner && series.TeacherID != callerID {
		return domain.ErrUnauthorizedLessonAction
	}

	switch scope {
	case domain.ScopeAllInSeries:
		return s.seriesRepo.Delete(ctx, series.ID)

	case domain.ScopeThisAndFollowing:
		if effectiveOccDate == nil {
			return s.seriesRepo.Delete(ctx, series.ID)
		}
		// Обрезаем серию: until_date = occurrenceDate - 1 day
		until := effectiveOccDate.AddDate(0, 0, -1)
		series.UntilDate = &until
		series.UpdatedAt = time.Now().UTC()
		return s.seriesRepo.Update(ctx, series)

	case domain.ScopeThisOnly:
		fallthrough
	default:
		if physicalLesson != nil {
			_, err := s.CancelLesson(ctx, physicalLesson.ID, callerID, callerRole, reason)
			return err
		}
		if effectiveOccDate == nil {
			return domain.ErrLessonNotFound
		}
		// Материализуем отмену как запись с status = cancelled
		hour, minute, _ := ParseTimeOfDay(series.StartTimeOfDay)
		loc := effectiveOccDate.Location()
		startTime := time.Date(effectiveOccDate.Year(), effectiveOccDate.Month(), effectiveOccDate.Day(), hour, minute, 0, 0, loc)
		endTime := startTime.Add(time.Duration(series.DurationMinutes) * time.Minute)

		now := time.Now().UTC()
		cancelledLesson := &domain.Lesson{
			ID:                uuid.New(),
			TeacherID:         series.TeacherID,
			ClientID:          series.ClientID,
			Title:             series.Title,
			ClassroomID:       series.ClassroomID,
			StartTime:         startTime,
			EndTime:           endTime,
			Format:            series.Format,
			LocationOrURL:     series.LocationOrURL,
			Status:            domain.StatusCancelled,
			Notes:             series.Notes,
			CancelReason:      reason,
			SeriesID:          &series.ID,
			OriginalStartTime: &startTime,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		return s.lessonRepo.Create(ctx, cancelledLesson)
	}
}

