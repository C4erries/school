package schedule

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/C4erries/school/backend/internal/domain"
)

// ListLessonsWithRecurring извлекает физические уроки и объединяет их с виртуальными слотами серий.
func (s *Service) ListLessonsWithRecurring(ctx context.Context, filter LessonFilter) ([]*domain.Lesson, error) {
	// 1. Извлекаем все физические уроки из БД
	physicalLessons, err := s.lessonRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list physical lessons: %w", err)
	}

	// Если не задано окно дат или нет репозитория серий или учителя — возвращаем физические уроки
	if filter.From == nil || filter.To == nil || s.seriesRepo == nil || filter.TeacherID == nil {
		return physicalLessons, nil
	}

	// 2. Индексируем физические исключения для серий (все уроки в окне независимо от статуса, включая cancelled):
	// key: seriesID.String() + "_" + originalStartTime.UTC().Format(time.RFC3339)
	seriesOverrides := make(map[string]*domain.Lesson)
	if filter.Status != nil {
		overrideFilter := LessonFilter{
			TeacherID: filter.TeacherID,
			From:      filter.From,
			To:        filter.To,
		}
		allInWindow, err := s.lessonRepo.List(ctx, overrideFilter)
		if err == nil {
			for _, l := range allInWindow {
				if l.SeriesID != nil && l.OriginalStartTime != nil {
					key := fmt.Sprintf("%s_%s", l.SeriesID.String(), l.OriginalStartTime.UTC().Format(time.RFC3339))
					seriesOverrides[key] = l
				}
			}
		}
	} else {
		for _, l := range physicalLessons {
			if l.SeriesID != nil && l.OriginalStartTime != nil {
				key := fmt.Sprintf("%s_%s", l.SeriesID.String(), l.OriginalStartTime.UTC().Format(time.RFC3339))
				seriesOverrides[key] = l
			}
		}
	}

	// 3. Извлекаем все серии преподавателя
	seriesList, err := s.seriesRepo.ListByTeacherID(ctx, *filter.TeacherID)
	if err != nil {
		return nil, fmt.Errorf("list series: %w", err)
	}

	// 4. Генерируем виртуальные слоты для каждой серии
	var virtualLessons []*domain.Lesson

	for _, series := range seriesList {
		if filter.ClientID != nil && series.ClientID != *filter.ClientID {
			continue
		}
		if filter.ClassroomID != nil {
			if series.ClassroomID == nil || *series.ClassroomID != *filter.ClassroomID {
				continue
			}
		}

		slots, err := GenerateOccurrences(series, *filter.From, *filter.To)
		if err != nil {
			continue
		}

		for _, slot := range slots {
			key := fmt.Sprintf("%s_%s", series.ID.String(), slot.OriginalStartTime.UTC().Format(time.RFC3339))
			// Если на этот слот есть физическая запись (перенос, завершение или отмена)
			if override, exists := seriesOverrides[key]; exists {
				// Физическая запись уже есть в physicalLessons. Виртуальный слот подавляем.
				_ = override
				continue
			}

			// Если фильтр требует конкретный статус, а виртуальный урок всегда scheduled
			if filter.Status != nil && *filter.Status != domain.StatusScheduled {
				continue
			}

			vLesson := BuildVirtualLesson(series, slot)
			virtualLessons = append(virtualLessons, vLesson)
		}
	}

	// 5. Объединяем физические и виртуальные уроки
	all := make([]*domain.Lesson, 0, len(physicalLessons)+len(virtualLessons))
	for _, l := range physicalLessons {
		// Если это физический урок со статусом cancelled и запрашивали статус scheduled — фильтруем
		if filter.Status != nil && l.Status != *filter.Status {
			continue
		}
		all = append(all, l)
	}
	all = append(all, virtualLessons...)

	// 6. Сортируем по StartTime возрастанию
	sort.Slice(all, func(i, j int) bool {
		return all[i].StartTime.Before(all[j].StartTime)
	})

	return all, nil
}
