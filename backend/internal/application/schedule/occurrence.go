package schedule

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
)

// OccurrenceSlot представляет сгенерированный временной слот для серии.
type OccurrenceSlot struct {
	StartTime         time.Time
	EndTime           time.Time
	OriginalStartTime time.Time
}

// ParseTimeOfDay парсит строку "15:04" или "15:04:00" в часы и минуты.
func ParseTimeOfDay(tod string) (hour int, minute int, err error) {
	parts := strings.Split(tod, ":")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("invalid time of day format: %s", tod)
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return h, m, nil
}

// GenerateOccurrences генерирует все слоты серии, попадающие в интервал [windowFrom, windowTo].
func GenerateOccurrences(series *domain.LessonSeries, windowFrom, windowTo time.Time) ([]OccurrenceSlot, error) {
	rule, err := domain.ParseWeeklyRRULE(series.RRULE)
	if err != nil {
		return nil, err
	}

	hour, minute, err := ParseTimeOfDay(series.StartTimeOfDay)
	if err != nil {
		return nil, err
	}

	// Начинаем от max(series.StartDate, windowFrom.Date)
	startLimit := series.StartDate
	if windowFrom.After(startLimit) {
		startLimit = time.Date(windowFrom.Year(), windowFrom.Month(), windowFrom.Day(), 0, 0, 0, 0, series.StartDate.Location())
	}

	// Заканчиваем min(series.UntilDate, windowTo.Date)
	endLimit := time.Date(windowTo.Year(), windowTo.Month(), windowTo.Day(), 23, 59, 59, 0, series.StartDate.Location())
	if series.UntilDate != nil && series.UntilDate.Before(endLimit) {
		endLimit = time.Date(series.UntilDate.Year(), series.UntilDate.Month(), series.UntilDate.Day(), 23, 59, 59, 0, series.StartDate.Location())
	}

	if startLimit.After(endLimit) {
		return nil, nil
	}

	dayMap := make(map[time.Weekday]bool, len(rule.DaysOfWeek))
	for _, d := range rule.DaysOfWeek {
		dayMap[d] = true
	}

	var slots []OccurrenceSlot
	loc := windowFrom.Location()
	if loc == nil {
		loc = time.UTC
	}

	cur := time.Date(startLimit.Year(), startLimit.Month(), startLimit.Day(), 0, 0, 0, 0, loc)
	untilCur := time.Date(endLimit.Year(), endLimit.Month(), endLimit.Day(), 23, 59, 59, 0, loc)

	for !cur.After(untilCur) {
		if dayMap[cur.Weekday()] {
			startTime := time.Date(cur.Year(), cur.Month(), cur.Day(), hour, minute, 0, 0, loc)
			endTime := startTime.Add(time.Duration(series.DurationMinutes) * time.Minute)

			// Проверяем пересечение с окном [windowFrom, windowTo]
			if !endTime.Before(windowFrom) && !startTime.After(windowTo) {
				slots = append(slots, OccurrenceSlot{
					StartTime:         startTime,
					EndTime:           endTime,
					OriginalStartTime: startTime,
				})
			}
		}
		cur = cur.AddDate(0, 0, 1)
	}

	return slots, nil
}

// VirtualLessonID вычисляет детерминированный UUID v5 для виртуального слота серии.
func VirtualLessonID(seriesID uuid.UUID, originalStartTime time.Time) uuid.UUID {
	return uuid.NewSHA1(seriesID, []byte(originalStartTime.UTC().Format(time.RFC3339)))
}

// BuildVirtualLesson создает виртуальный объект domain.Lesson для слота серии.
func BuildVirtualLesson(series *domain.LessonSeries, slot OccurrenceSlot) *domain.Lesson {
	return &domain.Lesson{
		ID:                VirtualLessonID(series.ID, slot.OriginalStartTime),
		TeacherID:         series.TeacherID,
		ClientID:          series.ClientID,
		Title:             series.Title,
		ClassroomID:       series.ClassroomID,
		StartTime:         slot.StartTime,
		EndTime:           slot.EndTime,
		Format:            series.Format,
		LocationOrURL:     series.LocationOrURL,
		Status:            domain.StatusScheduled,
		Notes:             series.Notes,
		SeriesID:          &series.ID,
		OriginalStartTime: &slot.OriginalStartTime,
		CreatedAt:         series.CreatedAt,
		UpdatedAt:         series.UpdatedAt,
	}
}

// FindVirtualLesson ищет виртуальный урок по его ID среди серий преподавателя.
func (s *Service) FindVirtualLesson(ctx context.Context, lessonID uuid.UUID, teacherID *uuid.UUID) (*domain.LessonSeries, *OccurrenceSlot, error) {
	if s.seriesRepo == nil || teacherID == nil {
		return nil, nil, domain.ErrLessonNotFound
	}

	seriesList, err := s.seriesRepo.ListByTeacherID(ctx, *teacherID)
	if err != nil {
		return nil, nil, err
	}

	for _, series := range seriesList {
		from := series.StartDate
		until := series.StartDate.AddDate(2, 0, 0)
		if series.UntilDate != nil && series.UntilDate.Before(until) {
			until = *series.UntilDate
		}
		slots, err := GenerateOccurrences(series, from, until)
		if err != nil {
			continue
		}
		for _, slot := range slots {
			if VirtualLessonID(series.ID, slot.OriginalStartTime) == lessonID {
				return series, &slot, nil
			}
		}
	}

	return nil, nil, domain.ErrLessonNotFound
}

// EnsurePhysicalLesson проверяет существование физического урока в БД.
// Если урок физический — возвращает его.
// Если урок виртуальный (слот серии) — материализует его в БД со статусом scheduled и возвращает созданный физический урок.
func (s *Service) EnsurePhysicalLesson(ctx context.Context, lessonID uuid.UUID, teacherID uuid.UUID) (*domain.Lesson, error) {
	lesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err == nil {
		if lesson.TeacherID != teacherID {
			return nil, domain.ErrUnauthorizedLessonAction
		}
		return lesson, nil
	}

	if !errors.Is(err, domain.ErrLessonNotFound) {
		return nil, fmt.Errorf("get lesson by id: %w", err)
	}

	// Урок не найден в БД — ищем среди виртуальных слотов регулярных серий
	series, slot, err := s.FindVirtualLesson(ctx, lessonID, &teacherID)
	if err != nil {
		return nil, domain.ErrLessonNotFound
	}

	now := time.Now().UTC()
	virtualLesson := BuildVirtualLesson(series, *slot)
	virtualLesson.CreatedAt = now
	virtualLesson.UpdatedAt = now

	if err := s.lessonRepo.Create(ctx, virtualLesson); err != nil {
		// В случае параллельного создания проверяем повторно
		if existing, getErr := s.lessonRepo.GetByID(ctx, lessonID); getErr == nil {
			return existing, nil
		}
		return nil, fmt.Errorf("materialize virtual lesson: %w", err)
	}

	return virtualLesson, nil
}


