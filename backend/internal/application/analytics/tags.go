package analytics

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
)

// TagStat статистика по тегу за период.
type TagStat struct {
	TagID          uuid.UUID
	TagName        string
	TagColor       *string
	StudentsCount  int
	CompletedHours float32
	GrossRevenue   float32
	NetIncome      float32
}

// GetTagStats группирует проведенные уроки (completed) по всем тегам клиентов (включая непартнерские с 0%).
func (s *Service) GetTagStats(ctx context.Context, teacherID uuid.UUID, from, to *time.Time) ([]TagStat, error) {
	startRange, endRange := resolveDateRange(from, to)

	statusCompleted := domain.StatusCompleted
	lessons, err := s.lessonRepo.List(ctx, schedule.LessonFilter{
		TeacherID: &teacherID,
		Status:    &statusCompleted,
		From:      &startRange,
		To:        &endRange,
	})
	if err != nil {
		return nil, fmt.Errorf("list lessons for tag stats: %w", err)
	}

	clients, err := s.clientRepo.ListByTeacherID(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list clients for tag stats: %w", err)
	}
	clientMap := make(map[uuid.UUID]*domain.Client, len(clients))
	for _, c := range clients {
		clientMap[c.ID] = c
	}

	type tagAccum struct {
		tagID          uuid.UUID
		tagName        string
		tagColor       *string
		clientIDs      map[uuid.UUID]bool
		completedHours float64
		grossRevenue   float64
		netIncome      float64
	}

	tagMap := make(map[uuid.UUID]*tagAccum)

	for _, l := range lessons {
		if l.StartTime.Before(startRange) || l.StartTime.After(endRange) {
			continue
		}
		client := clientMap[l.ClientID]
		if client == nil {
			continue
		}

		rev, comm, dur := calculateLessonRevenueAndCommission(l, client)
		net := rev - comm

		for _, t := range client.Tags {
			acc, ok := tagMap[t.ID]
			if !ok {
				var colorPtr *string
				if t.Color != "" {
					c := t.Color
					colorPtr = &c
				}
				acc = &tagAccum{
					tagID:     t.ID,
					tagName:   t.Name,
					tagColor:  colorPtr,
					clientIDs: make(map[uuid.UUID]bool),
				}
				tagMap[t.ID] = acc
			}
			acc.clientIDs[client.ID] = true
			acc.completedHours += dur
			acc.grossRevenue += rev
			acc.netIncome += net
		}
	}

	result := make([]TagStat, 0, len(tagMap))
	for _, acc := range tagMap {
		result = append(result, TagStat{
			TagID:          acc.tagID,
			TagName:        acc.tagName,
			TagColor:       acc.tagColor,
			StudentsCount:  len(acc.clientIDs),
			CompletedHours: roundFloat32(float32(acc.completedHours), 2),
			GrossRevenue:   roundFloat32(float32(acc.grossRevenue), 2),
			NetIncome:      roundFloat32(float32(acc.netIncome), 2),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].GrossRevenue == result[j].GrossRevenue {
			return result[i].TagName < result[j].TagName
		}
		return result[i].GrossRevenue > result[j].GrossRevenue
	})

	return result, nil
}

