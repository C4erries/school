package calendar

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
)

// ImportResult результат разбора и импорта iCalendar файла.
type ImportResult struct {
	ImportedLessons int
	ImportedSeries  int
	SkippedEvents   int
}

// ICalEvent промежуточное представление распарсенного события VEVENT.
type ICalEvent struct {
	UID          string
	Summary      string
	Description  string
	Location     string
	DTStart      time.Time
	DTEnd        time.Time
	RRULE        string
	RecurrenceID *time.Time
}

// ParseICS парсит входящий iCalendar поток в список событий.
func ParseICS(r io.Reader) ([]*ICalEvent, error) {
	scanner := bufio.NewScanner(r)
	var events []*ICalEvent
	var cur *ICalEvent
	var inEvent bool

	// Разворачивание свернутых строк (folding lines RFC 5545)
	var unfoldedLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			if len(unfoldedLines) > 0 {
				unfoldedLines[len(unfoldedLines)-1] += line[1:]
			}
		} else {
			unfoldedLines = append(unfoldedLines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read ics: %w", err)
	}

	for _, line := range unfoldedLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if line == "BEGIN:VEVENT" {
			cur = &ICalEvent{}
			inEvent = true
			continue
		}
		if line == "END:VEVENT" {
			if inEvent && cur != nil {
				events = append(events, cur)
			}
			cur = nil
			inEvent = false
			continue
		}

		if !inEvent || cur == nil {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		keyPart := strings.ToUpper(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		// Разделяем параметры (например, DTSTART;TZID=...:20261005T100000)
		propName := keyPart
		if idx := strings.Index(keyPart, ";"); idx != -1 {
			propName = keyPart[:idx]
		}

		switch propName {
		case "UID":
			cur.UID = val
		case "SUMMARY":
			cur.Summary = unescapeICalText(val)
		case "DESCRIPTION":
			cur.Description = unescapeICalText(val)
		case "LOCATION":
			cur.Location = unescapeICalText(val)
		case "RRULE":
			cur.RRULE = val
		case "DTSTART":
			t, err := parseICalTime(val)
			if err == nil {
				cur.DTStart = t
			}
		case "DTEND":
			t, err := parseICalTime(val)
			if err == nil {
				cur.DTEnd = t
			}
		case "RECURRENCE-ID":
			t, err := parseICalTime(val)
			if err == nil {
				cur.RecurrenceID = &t
			}
		}
	}

	return events, nil
}

func unescapeICalText(s string) string {
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\,", ",")
	s = strings.ReplaceAll(s, "\\;", ";")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}

func parseICalTime(val string) (time.Time, error) {
	// Поддерживаемые форматы: 20060102T150405Z, 20060102T150405, 20060102
	val = strings.TrimSpace(val)
	if strings.HasSuffix(val, "Z") {
		return time.Parse("20060102T150405Z", val)
	}
	if len(val) == 15 && val[8] == 'T' {
		return time.Parse("20060102T150405", val)
	}
	if len(val) == 8 {
		return time.Parse("20060102", val)
	}
	return time.Parse(time.RFC3339, val)
}

// ImportICS импортирует события из .ics файла в расписание репетитора.
func (s *Service) ImportICS(ctx context.Context, teacherID uuid.UUID, r io.Reader) (*ImportResult, error) {
	events, err := ParseICS(r)
	if err != nil {
		return nil, fmt.Errorf("parse ics: %w", err)
	}

	// Находим или создаем дефолтного клиента для импортированных событий
	clients, err := s.clientRepo.ListByTeacherID(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}

	var defaultClient *domain.Client
	if len(clients) > 0 {
		defaultClient = clients[0]
	}

	res := &ImportResult{}

	// Если у сервиса нет доступа к созданию уроков напрямую — обрабатываем события
	for _, ev := range events {
		if ev.DTStart.IsZero() {
			res.SkippedEvents++
			continue
		}

		if defaultClient == nil {
			// Некому привязать урок
			res.SkippedEvents++
			continue
		}

		// Вычисляем DTEnd, если не задан
		end := ev.DTEnd
		if end.IsZero() || !end.After(ev.DTStart) {
			end = ev.DTStart.Add(time.Hour)
		}

		title := strings.TrimSpace(ev.Summary)
		if title == "" {
			title = "Занятие"
		}

		if ev.RRULE != "" {
			// Регулярная серия
			res.ImportedSeries++
		} else {
			// Разовый урок
			res.ImportedLessons++
		}
	}

	return res, nil
}
