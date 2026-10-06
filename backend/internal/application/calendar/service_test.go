package calendar_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/application/calendar"
	"github.com/C4erries/school/backend/internal/application/crm"
	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if u := args.Get(0); u != nil {
		return u.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepo) GetByCalendarToken(ctx context.Context, token uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, token)
	if u := args.Get(0); u != nil {
		return u.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepo) RotateCalendarToken(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

type MockLessonRepo struct {
	mock.Mock
}

func (m *MockLessonRepo) List(ctx context.Context, filter schedule.LessonFilter) ([]*domain.Lesson, error) {
	args := m.Called(ctx, filter)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Lesson), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockClientRepo struct {
	mock.Mock
}

func (m *MockClientRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	args := m.Called(ctx, id)
	if c := args.Get(0); c != nil {
		return c.(*domain.Client), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockClientRepo) ListByTeacherID(ctx context.Context, teacherID uuid.UUID, filter ...crm.ClientFilter) ([]*domain.Client, error) {
	args := m.Called(ctx, teacherID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Client), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestCalendarService_GetSettingsAndRotateToken(t *testing.T) {
	userRepo := new(MockUserRepo)
	lessonRepo := new(MockLessonRepo)
	clientRepo := new(MockClientRepo)
	svc := calendar.NewService(userRepo, lessonRepo, clientRepo)

	ctx := context.Background()
	userID := uuid.New()
	initialToken := uuid.New()
	user := &domain.User{
		ID:            userID,
		CalendarToken: initialToken,
	}

	userRepo.On("GetByID", ctx, userID).Return(user, nil)
	tok, err := svc.GetSettings(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, initialToken, tok)

	newToken := uuid.New()
	userRepo.On("RotateCalendarToken", ctx, userID).Return(newToken, nil)
	rotatedTok, err := svc.RotateToken(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, newToken, rotatedTok)
}

func TestCalendarService_GetFeedICS(t *testing.T) {
	userRepo := new(MockUserRepo)
	lessonRepo := new(MockLessonRepo)
	clientRepo := new(MockClientRepo)
	svc := calendar.NewService(userRepo, lessonRepo, clientRepo)

	ctx := context.Background()
	token := uuid.New()
	teacherID := uuid.New()
	clientID := uuid.New()

	phone := "+79991234567"
	client := &domain.Client{
		ID:        clientID,
		TeacherID: teacherID,
		Name:      "Петр Петров",
		Phone:     &phone,
	}

	user := &domain.User{
		ID:            teacherID,
		CalendarToken: token,
	}

	loc := "https://telemost.yandex.ru/j/12345"
	notes := "Тема: Интегралы; дз проверено"
	lessonID := uuid.New()
	lesson := &domain.Lesson{
		ID:            lessonID,
		TeacherID:     teacherID,
		ClientID:      clientID,
		Title:         "Математика ЕГЭ",
		StartTime:     time.Date(2026, 10, 10, 15, 0, 0, 0, time.UTC),
		EndTime:       time.Date(2026, 10, 10, 16, 30, 0, 0, time.UTC),
		Format:        domain.FormatIndividual,
		LocationOrURL: loc,
		Status:        domain.StatusScheduled,
		Notes:         notes,
	}

	userRepo.On("GetByCalendarToken", ctx, token).Return(user, nil)
	lessonRepo.On("List", ctx, mock.Anything).Return([]*domain.Lesson{lesson}, nil)
	clientRepo.On("ListByTeacherID", ctx, teacherID).Return([]*domain.Client{client}, nil)

	data, err := svc.GetFeedICS(ctx, token)
	require.NoError(t, err)
	icsStr := string(data)

	// Validate RFC 5545 requirements
	assert.True(t, strings.HasPrefix(icsStr, "BEGIN:VCALENDAR\r\n"))
	assert.True(t, strings.HasSuffix(icsStr, "END:VCALENDAR\r\n"))
	assert.Contains(t, icsStr, "PRODID:-//School Tutor Assistant//RU\r\n")
	assert.Contains(t, icsStr, "VERSION:2.0\r\n")
	assert.Contains(t, icsStr, "BEGIN:VEVENT\r\n")
	assert.Contains(t, icsStr, "UID:lesson-"+lessonID.String()+"@school\r\n")
	assert.Contains(t, icsStr, "DTSTART:20261010T150000Z\r\n")
	assert.Contains(t, icsStr, "DTEND:20261010T163000Z\r\n")
	assert.Contains(t, icsStr, "SUMMARY:Математика ЕГЭ: Петр Петров\r\n")
	assert.Contains(t, icsStr, "STATUS:CONFIRMED\r\n")
	assert.Contains(t, icsStr, "LOCATION:https://telemost.yandex.ru/j/12345\r\n")
	assert.Contains(t, icsStr, "Ученик: Петр Петров")
	assert.Contains(t, icsStr, "Телефон: +79991234567")
	assert.Contains(t, icsStr, "END:VEVENT\r\n")
}

func TestCalendarService_ExportCalendarICS_CancelledLesson(t *testing.T) {
	userRepo := new(MockUserRepo)
	lessonRepo := new(MockLessonRepo)
	clientRepo := new(MockClientRepo)
	svc := calendar.NewService(userRepo, lessonRepo, clientRepo)

	ctx := context.Background()
	teacherID := uuid.New()
	clientID := uuid.New()

	client := &domain.Client{
		ID:        clientID,
		TeacherID: teacherID,
		Name:      "Ольга Сидорова",
	}

	user := &domain.User{
		ID: teacherID,
	}

	cancelReason := "Болезнь"
	lesson := &domain.Lesson{
		ID:           uuid.New(),
		TeacherID:    teacherID,
		ClientID:     clientID,
		Title:        "Химия",
		StartTime:    time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC),
		EndTime:      time.Date(2026, 10, 12, 11, 0, 0, 0, time.UTC),
		Format:       domain.FormatPair,
		Status:       domain.StatusCancelled,
		CancelReason: cancelReason,
	}

	userRepo.On("GetByID", ctx, teacherID).Return(user, nil)
	lessonRepo.On("List", ctx, mock.Anything).Return([]*domain.Lesson{lesson}, nil)
	clientRepo.On("ListByTeacherID", ctx, teacherID).Return([]*domain.Client{client}, nil)

	data, err := svc.ExportCalendarICS(ctx, teacherID, nil, nil)
	require.NoError(t, err)
	icsStr := string(data)

	assert.Contains(t, icsStr, "STATUS:CANCELLED\r\n")
	assert.Contains(t, icsStr, "Причина отмены: Болезнь")
	assert.Contains(t, icsStr, "Формат: Парный")
}
