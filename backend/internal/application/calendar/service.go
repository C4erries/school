package calendar

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/application/crm"
	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
)

// UserRepository определяет контракт доступа к пользователям для календаря.
type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByCalendarToken(ctx context.Context, token uuid.UUID) (*domain.User, error)
	RotateCalendarToken(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
}

// LessonRepository определяет контракт доступа к расписанию уроков.
type LessonRepository interface {
	List(ctx context.Context, filter schedule.LessonFilter) ([]*domain.Lesson, error)
}

// ClientRepository определяет контракт доступа к клиентам преподавателя.
type ClientRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error)
	ListByTeacherID(ctx context.Context, teacherID uuid.UUID, filter ...crm.ClientFilter) ([]*domain.Client, error)
}

// Service предоставляет бизнес-логику календарной синхронизации (RFC 5545 iCalendar).
type Service struct {
	userRepo   UserRepository
	lessonRepo LessonRepository
	clientRepo ClientRepository
}

// NewService создает новый экземпляр CalendarService.
func NewService(userRepo UserRepository, lessonRepo LessonRepository, clientRepo ClientRepository) *Service {
	return &Service{
		userRepo:   userRepo,
		lessonRepo: lessonRepo,
		clientRepo: clientRepo,
	}
}

// GetSettings возвращает персональный токен календаря пользователя.
func (s *Service) GetSettings(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return uuid.Nil, err
	}
	return u.CalendarToken, nil
}

// RotateToken перевыпускает токен календаря пользователя.
func (s *Service) RotateToken(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return s.userRepo.RotateCalendarToken(ctx, userID)
}

// GetFeedICS возвращает iCalendar поток для Google/Apple Calendar по секретному токену преподавателя.
// Охватывает уроки от 30 дней назад до 60 дней вперед.
func (s *Service) GetFeedICS(ctx context.Context, token uuid.UUID) ([]byte, error) {
	user, err := s.userRepo.GetByCalendarToken(ctx, token)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	now := time.Now().UTC()
	from := now.AddDate(0, 0, -30)
	to := now.AddDate(0, 0, 60)

	lessons, err := s.lessonRepo.List(ctx, schedule.LessonFilter{
		TeacherID: &user.ID,
		From:      &from,
		To:        &to,
	})
	if err != nil {
		return nil, fmt.Errorf("list feed lessons: %w", err)
	}

	clientMap := make(map[uuid.UUID]*domain.Client)
	if s.clientRepo != nil {
		clients, err := s.clientRepo.ListByTeacherID(ctx, user.ID)
		if err == nil {
			for _, c := range clients {
				clientMap[c.ID] = c
			}
		}
	}

	return s.GenerateICS(lessons, clientMap, now), nil
}

// ExportCalendarICS выполняет разовый экспорт расписания преподавателя за указанный период.
func (s *Service) ExportCalendarICS(ctx context.Context, userID uuid.UUID, from, to *time.Time) ([]byte, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	now := time.Now().UTC()
	var startTime, endTime time.Time

	if from != nil && !from.IsZero() {
		startTime = from.UTC()
	} else {
		startTime = now.AddDate(0, 0, -30)
	}

	if to != nil && !to.IsZero() {
		endTime = to.UTC()
	} else {
		endTime = now.AddDate(0, 0, 90)
	}

	lessons, err := s.lessonRepo.List(ctx, schedule.LessonFilter{
		TeacherID: &user.ID,
		From:      &startTime,
		To:        &endTime,
	})
	if err != nil {
		return nil, fmt.Errorf("list export lessons: %w", err)
	}

	clientMap := make(map[uuid.UUID]*domain.Client)
	if s.clientRepo != nil {
		clients, err := s.clientRepo.ListByTeacherID(ctx, user.ID)
		if err == nil {
			for _, c := range clients {
				clientMap[c.ID] = c
			}
		}
	}

	return s.GenerateICS(lessons, clientMap, now), nil
}

// escapeICalText экранирует спецсимволы RFC 5545 в текстовых полях.
func escapeICalText(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\r\n", "\\n")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// formatICalTime форматирует время в стандарте UTC iCalendar: 20060102T150405Z.
func formatICalTime(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}

// GenerateICS форматирует список занятий в валидный RFC 5545 iCalendar документ.
func (s *Service) GenerateICS(lessons []*domain.Lesson, clients map[uuid.UUID]*domain.Client, dtStamp time.Time) []byte {
	var buf bytes.Buffer

	// Календарный заголовок RFC 5545
	buf.WriteString("BEGIN:VCALENDAR\r\n")
	buf.WriteString("VERSION:2.0\r\n")
	buf.WriteString("PRODID:-//School Tutor Assistant//RU\r\n")
	buf.WriteString("CALSCALE:GREGORIAN\r\n")
	buf.WriteString("METHOD:PUBLISH\r\n")
	buf.WriteString("X-WR-CALNAME:Расписание уроков School\r\n")
	buf.WriteString("X-WR-TIMEZONE:UTC\r\n")

	stampStr := formatICalTime(dtStamp)

	for _, l := range lessons {
		client := clients[l.ClientID]

		buf.WriteString("BEGIN:VEVENT\r\n")
		buf.WriteString(fmt.Sprintf("UID:lesson-%s@school\r\n", l.ID.String()))
		buf.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", stampStr))
		buf.WriteString(fmt.Sprintf("DTSTART:%s\r\n", formatICalTime(l.StartTime)))
		buf.WriteString(fmt.Sprintf("DTEND:%s\r\n", formatICalTime(l.EndTime)))

		// SUMMARY
		summary := l.Title
		if client != nil {
			if summary == "" {
				summary = client.Name
			} else if !strings.Contains(summary, client.Name) {
				summary = fmt.Sprintf("%s: %s", summary, client.Name)
			}
		}
		if summary == "" {
			summary = "Урок"
		}
		buf.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICalText(summary)))

		// DESCRIPTION
		var descParts []string
		formatRu := "Индивидуальный"
		switch l.Format {
		case domain.FormatPair:
			formatRu = "Парный"
		case domain.FormatGroup:
			formatRu = "Групповой"
		}
		descParts = append(descParts, fmt.Sprintf("Формат: %s", formatRu))

		if client != nil {
			descParts = append(descParts, fmt.Sprintf("Ученик: %s", client.Name))
			if client.Phone != nil && *client.Phone != "" {
				descParts = append(descParts, fmt.Sprintf("Телефон: %s", *client.Phone))
			}
		}
		if l.Notes != "" {
			descParts = append(descParts, fmt.Sprintf("Заметки: %s", l.Notes))
		}
		if l.CancelReason != "" {
			descParts = append(descParts, fmt.Sprintf("Причина отмены: %s", l.CancelReason))
		}
		description := strings.Join(descParts, "\n")
		buf.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICalText(description)))

		// LOCATION
		if l.LocationOrURL != "" {
			buf.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeICalText(l.LocationOrURL)))
		}

		// STATUS
		status := "CONFIRMED"
		if l.Status == domain.StatusCancelled {
			status = "CANCELLED"
		}
		buf.WriteString(fmt.Sprintf("STATUS:%s\r\n", status))

		buf.WriteString("END:VEVENT\r\n")
	}

	buf.WriteString("END:VCALENDAR\r\n")

	return buf.Bytes()
}
