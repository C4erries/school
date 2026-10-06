package httpserver_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/application/analytics"
	"github.com/C4erries/school/backend/internal/application/auth"
	"github.com/C4erries/school/backend/internal/application/calendar"
	"github.com/C4erries/school/backend/internal/application/crm"
	"github.com/C4erries/school/backend/internal/application/dashboard"
	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/security"
)

type MockAnalyticsLessonRepo struct {
	mock.Mock
}

func (m *MockAnalyticsLessonRepo) List(ctx context.Context, filter schedule.LessonFilter) ([]*domain.Lesson, error) {
	args := m.Called(ctx, filter)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Lesson), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockAnalyticsClientRepo struct {
	mock.Mock
}

func (m *MockAnalyticsClientRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	args := m.Called(ctx, id)
	if c := args.Get(0); c != nil {
		return c.(*domain.Client), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAnalyticsClientRepo) ListByTeacherID(ctx context.Context, teacherID uuid.UUID, filter ...crm.ClientFilter) ([]*domain.Client, error) {
	args := m.Called(ctx, teacherID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Client), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockCalendarUserRepo struct {
	mock.Mock
}

func (m *MockCalendarUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if u := args.Get(0); u != nil {
		return u.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCalendarUserRepo) GetByCalendarToken(ctx context.Context, token uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, token)
	if u := args.Get(0); u != nil {
		return u.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCalendarUserRepo) RotateCalendarToken(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func setupAnalyticsAndCalendarHandler(t *testing.T) (
	*http.ServeMux,
	*MockAnalyticsLessonRepo,
	*MockAnalyticsClientRepo,
	*MockCalendarUserRepo,
	uuid.UUID,
	uuid.UUID,
) {
	tokenManager := new(MockTokenManager)
	lessonRepo := new(MockAnalyticsLessonRepo)
	clientRepo := new(MockAnalyticsClientRepo)
	userRepo := new(MockCalendarUserRepo)

	analyticsSvc := analytics.NewService(lessonRepo, clientRepo)
	calendarSvc := calendar.NewService(userRepo, lessonRepo, clientRepo)

	authService := auth.NewService(nil, nil, nil, nil)
	scheduleService := schedule.NewService(nil, nil, nil, nil)
	crmService := crm.NewService(nil, nil, nil, nil)
	dashboardService := dashboard.NewService(nil, nil)

	handler := httpserver.NewAPIHandler(
		authService,
		scheduleService,
		crmService,
		dashboardService,
		tokenManager,
		"v1",
		analyticsSvc,
		calendarSvc,
	)

	mux := httpserver.BuildMux(handler, nil)

	teacherID := uuid.New()
	calendarToken := uuid.New()

	tokenManager.On("ValidateAccessToken", "teacher_token").Return(&security.UserClaims{
		UserID: teacherID,
		Email:  "teacher@school.ru",
		Role:   domain.RoleTeacher,
	}, nil).Maybe()

	return mux, lessonRepo, clientRepo, userRepo, teacherID, calendarToken
}

func TestAnalyticsAndCalendarHandler(t *testing.T) {
	mux, lessonRepo, clientRepo, userRepo, teacherID, calendarToken := setupAnalyticsAndCalendarHandler(t)

	t.Run("GET /api/v1/analytics/overview - success", func(t *testing.T) {
		lessonRepo.On("List", mock.Anything, mock.Anything).Return([]*domain.Lesson{}, nil).Once()
		clientRepo.On("ListByTeacherID", mock.Anything, teacherID).Return([]*domain.Client{}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/overview", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp generated.AnalyticsOverviewResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, 0, resp.TotalLessons)
	})

	t.Run("GET /api/v1/analytics/dynamics - success", func(t *testing.T) {
		lessonRepo.On("List", mock.Anything, mock.Anything).Return([]*domain.Lesson{}, nil).Once()
		clientRepo.On("ListByTeacherID", mock.Anything, teacherID).Return([]*domain.Client{}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/dynamics?interval=month", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp []generated.AnalyticsDynamicsPoint
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.NotEmpty(t, resp)
	})

	t.Run("GET /api/v1/analytics/formats - success", func(t *testing.T) {
		lessonRepo.On("List", mock.Anything, mock.Anything).Return([]*domain.Lesson{}, nil).Once()
		clientRepo.On("ListByTeacherID", mock.Anything, teacherID).Return([]*domain.Client{}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/formats", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp []generated.AnalyticsFormatStat
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Len(t, resp, 3)
	})

	t.Run("GET /api/v1/analytics/clients - success", func(t *testing.T) {
		lessonRepo.On("List", mock.Anything, mock.Anything).Return([]*domain.Lesson{}, nil).Once()
		clientRepo.On("ListByTeacherID", mock.Anything, teacherID).Return([]*domain.Client{}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/clients?sort=hours&limit=10", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp []generated.AnalyticsClientStat
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Empty(t, resp)
	})

	t.Run("GET /api/v1/integrations/calendar/settings - success", func(t *testing.T) {
		user := &domain.User{
			ID:            teacherID,
			CalendarToken: calendarToken,
		}
		userRepo.On("GetByID", mock.Anything, teacherID).Return(user, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/calendar/settings", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp generated.CalendarSettingsResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, calendarToken, resp.CalendarToken)
		assert.Contains(t, resp.FeedUrl, calendarToken.String())
		assert.Contains(t, resp.WebcalUrl, calendarToken.String())
	})

	t.Run("POST /api/v1/integrations/calendar/rotate-token - success", func(t *testing.T) {
		newTok := uuid.New()
		userRepo.On("RotateCalendarToken", mock.Anything, teacherID).Return(newTok, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/calendar/rotate-token", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp generated.CalendarSettingsResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, newTok, resp.CalendarToken)
		assert.Contains(t, resp.FeedUrl, newTok.String())
	})

	t.Run("GET /api/v1/integrations/calendar/export - success", func(t *testing.T) {
		user := &domain.User{
			ID: teacherID,
		}
		userRepo.On("GetByID", mock.Anything, teacherID).Return(user, nil).Once()
		lessonRepo.On("List", mock.Anything, mock.Anything).Return([]*domain.Lesson{}, nil).Once()
		clientRepo.On("ListByTeacherID", mock.Anything, teacherID).Return([]*domain.Client{}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/calendar/export", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "text/calendar; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Body.String(), "BEGIN:VCALENDAR\r\n")
		assert.Contains(t, rr.Body.String(), "END:VCALENDAR\r\n")
	})

	t.Run("GET /api/v1/integrations/calendar/feed.ics - public access via query token", func(t *testing.T) {
		feedToken := uuid.New()
		user := &domain.User{
			ID:            teacherID,
			CalendarToken: feedToken,
		}
		userRepo.On("GetByCalendarToken", mock.Anything, feedToken).Return(user, nil).Once()
		lessonRepo.On("List", mock.Anything, mock.Anything).Return([]*domain.Lesson{}, nil).Once()
		clientRepo.On("ListByTeacherID", mock.Anything, teacherID).Return([]*domain.Client{}, nil).Once()

		// Request WITHOUT Authorization header! Public access!
		req := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/calendar/feed.ics?token="+feedToken.String(), nil)
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "text/calendar; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Body.String(), "BEGIN:VCALENDAR\r\n")
		assert.Contains(t, rr.Body.String(), "END:VCALENDAR\r\n")
	})

	t.Run("GET /api/v1/analytics/forecast - success", func(t *testing.T) {
		lessonRepo.On("List", mock.Anything, mock.Anything).Return([]*domain.Lesson{}, nil).Once()
		clientRepo.On("ListByTeacherID", mock.Anything, teacherID).Return([]*domain.Client{}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/forecast", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp generated.AnalyticsForecastResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, 0, resp.ScheduledLessons)
		assert.Equal(t, float32(0), resp.GrossPotentialRevenue)
		assert.Len(t, resp.ByFormat, 3)
	})

	t.Run("GET /api/v1/analytics/tags - success", func(t *testing.T) {
		lessonRepo.On("List", mock.Anything, mock.Anything).Return([]*domain.Lesson{}, nil).Once()
		clientRepo.On("ListByTeacherID", mock.Anything, teacherID).Return([]*domain.Client{}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/tags", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp []generated.TagStatResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Empty(t, resp)
	})
}
