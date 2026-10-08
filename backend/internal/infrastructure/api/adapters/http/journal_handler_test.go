package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/application/journal"
	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
	httpadapter "github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/security"
	"github.com/C4erries/school/backend/internal/infrastructure/api/config"
)

type mockJournalRepoForHTTP struct {
	mock.Mock
}

func (m *mockJournalRepoForHTTP) GetByLessonID(ctx context.Context, lessonID uuid.UUID) (*domain.LessonJournal, error) {
	args := m.Called(ctx, lessonID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LessonJournal), args.Error(1)
}

func (m *mockJournalRepoForHTTP) Upsert(ctx context.Context, j *domain.LessonJournal) error {
	args := m.Called(ctx, j)
	return args.Error(0)
}

func (m *mockJournalRepoForHTTP) ListByClientID(ctx context.Context, clientID, teacherID uuid.UUID) ([]*domain.LessonJournal, error) {
	args := m.Called(ctx, clientID, teacherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LessonJournal), args.Error(1)
}

type mockHomeworkRepoForHTTP struct {
	mock.Mock
}

func (m *mockHomeworkRepoForHTTP) Create(ctx context.Context, hw *domain.HomeworkAssignment) error {
	args := m.Called(ctx, hw)
	return args.Error(0)
}

func (m *mockHomeworkRepoForHTTP) GetByID(ctx context.Context, id uuid.UUID) (*domain.HomeworkAssignment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.HomeworkAssignment), args.Error(1)
}

func (m *mockHomeworkRepoForHTTP) Update(ctx context.Context, hw *domain.HomeworkAssignment) error {
	args := m.Called(ctx, hw)
	return args.Error(0)
}

func (m *mockHomeworkRepoForHTTP) Delete(ctx context.Context, id, teacherID uuid.UUID) error {
	args := m.Called(ctx, id, teacherID)
	return args.Error(0)
}

func (m *mockHomeworkRepoForHTTP) ListByClientID(ctx context.Context, clientID, teacherID uuid.UUID, statusFilter *domain.HomeworkStatus) ([]*domain.HomeworkAssignment, error) {
	args := m.Called(ctx, clientID, teacherID, statusFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.HomeworkAssignment), args.Error(1)
}

func (m *mockHomeworkRepoForHTTP) ListByAssignedLessonID(ctx context.Context, lessonID uuid.UUID) ([]*domain.HomeworkAssignment, error) {
	args := m.Called(ctx, lessonID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.HomeworkAssignment), args.Error(1)
}

func (m *mockHomeworkRepoForHTTP) ListDueByLessonOrDate(ctx context.Context, clientID uuid.UUID, targetDate time.Time) ([]*domain.HomeworkAssignment, error) {
	args := m.Called(ctx, clientID, targetDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.HomeworkAssignment), args.Error(1)
}

type mockScheduleProvForHTTP struct {
	mock.Mock
}

func (m *mockScheduleProvForHTTP) GetLesson(ctx context.Context, lessonID uuid.UUID) (*domain.Lesson, error) {
	args := m.Called(ctx, lessonID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Lesson), args.Error(1)
}

func (m *mockScheduleProvForHTTP) FindVirtualLesson(ctx context.Context, lessonID uuid.UUID, teacherID *uuid.UUID) (*domain.LessonSeries, *schedule.OccurrenceSlot, error) {
	args := m.Called(ctx, lessonID, teacherID)
	var series *domain.LessonSeries
	var slot *schedule.OccurrenceSlot
	if args.Get(0) != nil {
		series = args.Get(0).(*domain.LessonSeries)
	}
	if args.Get(1) != nil {
		slot = args.Get(1).(*schedule.OccurrenceSlot)
	}
	return series, slot, args.Error(2)
}

func (m *mockScheduleProvForHTTP) EnsurePhysicalLesson(ctx context.Context, lessonID uuid.UUID, teacherID uuid.UUID) (*domain.Lesson, error) {
	args := m.Called(ctx, lessonID, teacherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Lesson), args.Error(1)
}

func (m *mockScheduleProvForHTTP) GetUpcomingLesson(ctx context.Context, clientID, teacherID uuid.UUID) (*domain.Lesson, error) {
	args := m.Called(ctx, clientID, teacherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Lesson), args.Error(1)
}

type mockClientProvForHTTP struct {
	mock.Mock
}

func (m *mockClientProvForHTTP) GetClient(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Client), args.Error(1)
}

type mockNotesRepoForHTTP struct {
	mock.Mock
}

func (m *mockNotesRepoForHTTP) Create(ctx context.Context, note *domain.ClientNote) error {
	args := m.Called(ctx, note)
	return args.Error(0)
}

func (m *mockNotesRepoForHTTP) GetByID(ctx context.Context, id uuid.UUID) (*domain.ClientNote, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ClientNote), args.Error(1)
}

func (m *mockNotesRepoForHTTP) Delete(ctx context.Context, id, teacherID uuid.UUID) error {
	args := m.Called(ctx, id, teacherID)
	return args.Error(0)
}

func (m *mockNotesRepoForHTTP) ListByClientID(ctx context.Context, clientID, teacherID uuid.UUID) ([]*domain.ClientNote, error) {
	args := m.Called(ctx, clientID, teacherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ClientNote), args.Error(1)
}

func setupJournalTest(t *testing.T) (
	*httpadapter.Server,
	*mockJournalRepoForHTTP,
	*mockHomeworkRepoForHTTP,
	*mockScheduleProvForHTTP,
	*mockClientProvForHTTP,
	*mockNotesRepoForHTTP,
	uuid.UUID,
	string,
) {
	t.Helper()
	teacherID := uuid.New()
	tokenManager := new(MockTokenManager)
	token := "teacher_token"
	tokenManager.On("ValidateAccessToken", token).Return(&security.UserClaims{
		UserID: teacherID,
		Email:  "teacher@school.ru",
		Role:   domain.RoleTeacher,
	}, nil).Maybe()

	jRepo := new(mockJournalRepoForHTTP)
	hwRepo := new(mockHomeworkRepoForHTTP)
	sProv := new(mockScheduleProvForHTTP)
	cProv := new(mockClientProvForHTTP)
	nRepo := new(mockNotesRepoForHTTP)

	journalSvc := journal.NewService(jRepo, hwRepo, sProv, cProv, nRepo)

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	cfg := &config.Config{
		App: config.AppConfig{
			Port:    8080,
			Version: "v1",
		},
	}
	server := httpadapter.NewServer(cfg, logger, nil, nil, nil, nil, tokenManager, "v1", journalSvc)

	return server, jRepo, hwRepo, sProv, cProv, nRepo, teacherID, token
}

func TestJournalAndHomeworkHandler_Endpoints(t *testing.T) {
	server, jRepo, hwRepo, sProv, cProv, nRepo, teacherID, token := setupJournalTest(t)
	e := server.Echo()
	clientID := uuid.New()
	lessonID := uuid.New()
	now := time.Now().UTC()

	t.Run("GET /api/v1/schedule/lessons/{id}/journal - success", func(t *testing.T) {
		physLesson := &domain.Lesson{
			ID:        lessonID,
			TeacherID: teacherID,
			ClientID:  clientID,
			StartTime: now,
		}
		score := 5
		journalObj := &domain.LessonJournal{
			ID:               uuid.New(),
			LessonID:         lessonID,
			ClientID:         clientID,
			TeacherID:        teacherID,
			Topic:            "Теорема Пифагора",
			Notes:            "Все понятно",
			PerformanceScore: &score,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		sProv.On("GetLesson", mock.Anything, lessonID).Return(physLesson, nil).Once()
		jRepo.On("GetByLessonID", mock.Anything, lessonID).Return(journalObj, nil).Once()
		hwRepo.On("ListByAssignedLessonID", mock.Anything, lessonID).Return([]*domain.HomeworkAssignment{}, nil).Once()
		hwRepo.On("ListDueByLessonOrDate", mock.Anything, clientID, now).Return([]*domain.HomeworkAssignment{}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/schedule/lessons/"+lessonID.String()+"/journal", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		e.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var bundle generated.LessonJournalBundleResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &bundle))
		require.NotNil(t, bundle.Journal)
		assert.Equal(t, "Теорема Пифагора", bundle.Journal.Topic)
	})

	t.Run("PUT /api/v1/schedule/lessons/{id}/journal - success", func(t *testing.T) {
		physLesson := &domain.Lesson{
			ID:        lessonID,
			TeacherID: teacherID,
			ClientID:  clientID,
		}
		sProv.On("EnsurePhysicalLesson", mock.Anything, lessonID, teacherID).Return(physLesson, nil).Once()
		jRepo.On("Upsert", mock.Anything, mock.AnythingOfType("*domain.LessonJournal")).Return(nil).Once()

		score := 4
		body, _ := json.Marshal(generated.UpsertLessonJournalRequest{
			Topic:            "Квадратные уравнения",
			PerformanceScore: &score,
		})

		req := httptest.NewRequest(http.MethodPut, "/api/v1/schedule/lessons/"+lessonID.String()+"/journal", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		e.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp generated.LessonJournalResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "Квадратные уравнения", resp.Topic)
		require.NotNil(t, resp.PerformanceScore)
		assert.Equal(t, 4, *resp.PerformanceScore)
	})

	t.Run("GET /api/v1/crm/clients/{id}/journal - success", func(t *testing.T) {
		cProv.On("GetClient", mock.Anything, clientID).Return(&domain.Client{
			ID:        clientID,
			TeacherID: teacherID,
		}, nil).Once()
		jRepo.On("ListByClientID", mock.Anything, clientID, teacherID).Return([]*domain.LessonJournal{}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crm/clients/"+clientID.String()+"/journal", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		e.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp []generated.LessonJournalResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Empty(t, resp)
	})

	t.Run("POST /api/v1/crm/clients/{id}/homework - success", func(t *testing.T) {
		cProv.On("GetClient", mock.Anything, clientID).Return(&domain.Client{
			ID:        clientID,
			TeacherID: teacherID,
		}, nil).Once()
		hwRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.HomeworkAssignment")).Return(nil).Once()

		body, _ := json.Marshal(generated.CreateHomeworkRequest{
			Title: "Номера 10-15",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/crm/clients/"+clientID.String()+"/homework", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		e.ServeHTTP(rr, req)

		require.Equal(t, http.StatusCreated, rr.Code)
		var resp generated.HomeworkAssignmentResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "Номера 10-15", resp.Title)
		assert.Equal(t, generated.HomeworkAssignmentResponseStatusAssigned, resp.Status)
	})

	t.Run("PATCH /api/v1/homework/{id} - success", func(t *testing.T) {
		hwID := uuid.New()
		existingHW := &domain.HomeworkAssignment{
			ID:        hwID,
			ClientID:  clientID,
			TeacherID: teacherID,
			Title:     "Номера 10-15",
			Status:    domain.StatusHomeworkAssigned,
		}

		hwRepo.On("GetByID", mock.Anything, hwID).Return(existingHW, nil).Once()
		hwRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.HomeworkAssignment")).Return(nil).Once()

		revNotes := "Отлично решено"
		body, _ := json.Marshal(generated.UpdateHomeworkStatusRequest{
			Status:      generated.UpdateHomeworkStatusRequestStatusCompleted,
			ReviewNotes: &revNotes,
		})

		req := httptest.NewRequest(http.MethodPatch, "/api/v1/homework/"+hwID.String(), bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		e.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp generated.HomeworkAssignmentResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, generated.HomeworkAssignmentResponseStatusCompleted, resp.Status)
		require.NotNil(t, resp.ReviewNotes)
		assert.Equal(t, "Отлично решено", *resp.ReviewNotes)
	})

	t.Run("DELETE /api/v1/homework/{id} - success", func(t *testing.T) {
		hwID := uuid.New()
		hwRepo.On("Delete", mock.Anything, hwID, teacherID).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/homework/"+hwID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		e.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("POST /api/v1/crm/clients/{id}/notes - success", func(t *testing.T) {
		cProv.On("GetClient", mock.Anything, clientID).Return(&domain.Client{
			ID:        clientID,
			TeacherID: teacherID,
		}, nil).Once()
		nRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.ClientNote")).Return(nil).Once()

		body, _ := json.Marshal(generated.CreateClientNoteRequest{
			Content: "Позвонили родители, обсудили успеваемость",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/crm/clients/"+clientID.String()+"/notes", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		e.ServeHTTP(rr, req)

		require.Equal(t, http.StatusCreated, rr.Code)
		var resp generated.ClientNoteResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "Позвонили родители, обсудили успеваемость", resp.Content)
		assert.Equal(t, clientID, resp.ClientId)
		assert.Equal(t, teacherID, resp.TeacherId)
	})

	t.Run("DELETE /api/v1/crm/clients/notes/{id} - success", func(t *testing.T) {
		noteID := uuid.New()
		nRepo.On("Delete", mock.Anything, noteID, teacherID).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/crm/clients/notes/"+noteID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		e.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("GET /api/v1/crm/clients/{id}/stream - success", func(t *testing.T) {
		cProv.On("GetClient", mock.Anything, clientID).Return(&domain.Client{
			ID:        clientID,
			TeacherID: teacherID,
		}, nil).Once()

		jRepo.On("ListByClientID", mock.Anything, clientID, teacherID).Return([]*domain.LessonJournal{}, nil).Once()
		nRepo.On("ListByClientID", mock.Anything, clientID, teacherID).Return([]*domain.ClientNote{
			{
				ID:        uuid.New(),
				ClientID:  clientID,
				TeacherID: teacherID,
				Content:   "Быстрая заметка",
				CreatedAt: now,
			},
		}, nil).Once()

		upcoming := &domain.Lesson{
			ID:        uuid.New(),
			TeacherID: teacherID,
			ClientID:  clientID,
			Title:     "Ближайший урок",
			StartTime: now.Add(24 * time.Hour),
			EndTime:   now.Add(25 * time.Hour),
			Format:    domain.FormatIndividual,
			Notes:     "Тема следующего урока",
			Status:    domain.StatusScheduled,
		}
		sProv.On("GetUpcomingLesson", mock.Anything, clientID, teacherID).Return(upcoming, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crm/clients/"+clientID.String()+"/stream", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		e.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp generated.StudyStreamResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, clientID, resp.ClientId)
		require.Len(t, resp.Items, 1)
		assert.Equal(t, generated.Note, resp.Items[0].Type)
		assert.Equal(t, "Быстрая заметка", resp.Items[0].Note.Content)
		require.NotNil(t, resp.UpcomingLesson)
		assert.Equal(t, "Ближайший урок", resp.UpcomingLesson.Title)
		require.NotNil(t, resp.UpcomingLesson.Topic)
		assert.Equal(t, "Тема следующего урока", *resp.UpcomingLesson.Topic)
	})
}
