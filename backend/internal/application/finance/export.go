package finance

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
)

// --- CSV Экспорт данных (UTF-8 BOM) ---

const utf8BOM = "\xEF\xBB\xBF"

// ExportClientsCSV формирует CSV файл реестра учеников с кодировкой UTF-8 BOM.
func (s *Service) ExportClientsCSV(ctx context.Context, teacherID uuid.UUID) ([]byte, error) {
	clients, err := s.clientRepo.ListByTeacherID(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString(utf8BOM)
	w := csv.NewWriter(&buf)

	header := []string{
		"ID",
		"Имя",
		"Телефон",
		"Базовая ставка (₽)",
		"Ставка индивид. (₽)",
		"Ставка пара (₽)",
		"Ставка группа (₽)",
		"Баланс индивид. (ч)",
		"Баланс пара (ч)",
		"Баланс группа (ч)",
		"Общий баланс (ч)",
		"Теги",
		"В архиве",
		"Дата создания",
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	for _, c := range clients {
		phone := ""
		if c.Phone != nil {
			phone = *c.Phone
		}
		ratePair := ""
		if c.RatePair != nil {
			ratePair = fmt.Sprintf("%.2f", *c.RatePair)
		}
		rateGroup := ""
		if c.RateGroup != nil {
			rateGroup = fmt.Sprintf("%.2f", *c.RateGroup)
		}

		tagNames := make([]string, 0, len(c.Tags))
		for _, t := range c.Tags {
			tagNames = append(tagNames, t.Name)
		}

		isArchived := "Нет"
		if c.IsArchived {
			isArchived = "Да"
		}

		row := []string{
			c.ID.String(),
			c.Name,
			phone,
			fmt.Sprintf("%.2f", c.BaseRate),
			fmt.Sprintf("%.2f", c.RateIndividual),
			ratePair,
			rateGroup,
			strconv.FormatFloat(c.Balances.IndividualHours, 'f', 2, 64),
			strconv.FormatFloat(c.Balances.PairHours, 'f', 2, 64),
			strconv.FormatFloat(c.Balances.GroupHours, 'f', 2, 64),
			strconv.FormatFloat(c.Balances.TotalHours, 'f', 2, 64),
			strings.Join(tagNames, ", "),
			isArchived,
			c.CreatedAt.Format("2006-01-02 15:04"),
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ExportLessonsCSV формирует CSV файл расписания уроков с фильтром по датам и UTF-8 BOM.
func (s *Service) ExportLessonsCSV(ctx context.Context, teacherID uuid.UUID, from, to *time.Time) ([]byte, error) {
	lessons, err := s.lessonRepo.List(ctx, schedule.LessonFilter{
		TeacherID: &teacherID,
		From:      from,
		To:        to,
	})
	if err != nil {
		return nil, fmt.Errorf("list lessons: %w", err)
	}

	clientNames := make(map[uuid.UUID]string)
	if s.clientRepo != nil {
		if clients, err := s.clientRepo.ListByTeacherID(ctx, teacherID); err == nil {
			for _, c := range clients {
				clientNames[c.ID] = c.Name
			}
		}
	}

	var buf bytes.Buffer
	buf.WriteString(utf8BOM)
	w := csv.NewWriter(&buf)

	header := []string{
		"ID",
		"Дата начала",
		"Дата окончания",
		"Тема",
		"Формат",
		"Ученик",
		"Статус",
		"Локация / Ссылка",
		"Заметки",
		"Причина отмены",
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	for _, l := range lessons {
		clientName := clientNames[l.ClientID]
		if clientName == "" && s.clientRepo != nil {
			if c, err := s.clientRepo.GetByID(ctx, l.ClientID); err == nil && c != nil {
				clientName = c.Name
				clientNames[l.ClientID] = clientName
			}
		}

		formatName := string(l.Format)
		switch l.Format {
		case domain.FormatIndividual:
			formatName = "Индивидуальное"
		case domain.FormatPair:
			formatName = "Парное"
		case domain.FormatGroup:
			formatName = "Мини-группа"
		}

		statusName := string(l.Status)
		switch l.Status {
		case domain.StatusScheduled:
			statusName = "Запланирован"
		case domain.StatusCompleted:
			statusName = "Завершен"
		case domain.StatusCancelled:
			statusName = "Отменен"
		}

		row := []string{
			l.ID.String(),
			l.StartTime.Format("2006-01-02 15:04"),
			l.EndTime.Format("2006-01-02 15:04"),
			l.Title,
			formatName,
			clientName,
			statusName,
			l.LocationOrURL,
			l.Notes,
			l.CancelReason,
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ExportPaymentsCSV формирует CSV файл журнала оплат с фильтром по датам и UTF-8 BOM.
func (s *Service) ExportPaymentsCSV(ctx context.Context, teacherID uuid.UUID, from, to *time.Time) ([]byte, error) {
	payments, err := s.paymentRepo.List(ctx, PaymentFilter{
		TeacherID: &teacherID,
		From:      from,
		To:        to,
	})
	if err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}

	clientNames := make(map[uuid.UUID]string)
	if s.clientRepo != nil {
		if clients, err := s.clientRepo.ListByTeacherID(ctx, teacherID); err == nil {
			for _, c := range clients {
				clientNames[c.ID] = c.Name
			}
		}
	}

	var buf bytes.Buffer
	buf.WriteString(utf8BOM)
	w := csv.NewWriter(&buf)

	header := []string{
		"ID",
		"Дата оплаты",
		"Ученик",
		"Сумма (₽)",
		"Часы",
		"Формат",
		"Способ оплаты",
		"Заметки",
		"Дата создания",
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	for _, p := range payments {
		clientName := clientNames[p.ClientID]
		if clientName == "" && s.clientRepo != nil {
			if c, err := s.clientRepo.GetByID(ctx, p.ClientID); err == nil && c != nil {
				clientName = c.Name
				clientNames[p.ClientID] = clientName
			}
		}

		formatName := string(p.Format)
		switch p.Format {
		case domain.SubscriptionFormatIndividual:
			formatName = "Индивидуальное"
		case domain.SubscriptionFormatPair:
			formatName = "Парное"
		case domain.SubscriptionFormatGroup:
			formatName = "Мини-группа"
		}

		methodName := string(p.PaymentMethod)
		switch p.PaymentMethod {
		case domain.PaymentMethodTransfer:
			methodName = "Перевод / СБП"
		case domain.PaymentMethodCash:
			methodName = "Наличные"
		case domain.PaymentMethodCard:
			methodName = "Банковская карта"
		case domain.PaymentMethodOther:
			methodName = "Другое"
		}

		row := []string{
			p.ID.String(),
			p.PaidAt.Format("2006-01-02 15:04"),
			clientName,
			fmt.Sprintf("%.2f", p.Amount),
			strconv.FormatFloat(p.Hours, 'f', 2, 64),
			formatName,
			methodName,
			p.Notes,
			p.CreatedAt.Format("2006-01-02 15:04"),
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

