package httpserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/security"
)

type MockClassroomRepo struct {
	mock.Mock
}

func (m *MockClassroomRepo) Create(ctx context.Context, c *domain.Classroom) error {
	return m.Called(ctx, c).Error(0)
}
func (m *MockClassroomRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Classroom, error) {
	args := m.Called(ctx, id)
	if c := args.Get(0); c != nil {
		return c.(*domain.Classroom), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockClassroomRepo) List(ctx context.Context) ([]*domain.Classroom, error) {
	args := m.Called(ctx)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Classroom), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockClassroomRepo) Update(ctx context.Context, c *domain.Classroom) error {
	return m.Called(ctx, c).Error(0)
}
func (m *MockClassroomRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
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

type MockSubscriptionRepo struct {
	mock.Mock
}

func (m *MockSubscriptionRepo) GetByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.ClientSubscription, error) {
	args := m.Called(ctx, clientID)
	if subs := args.Get(0); subs != nil {
		return subs.([]*domain.ClientSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSubscriptionRepo) Update(ctx context.Context, sub *domain.ClientSubscription) error {
	return m.Called(ctx, sub).Error(0)
}

type MockLessonRepo struct {
	mock.Mock
}

func (m *MockLessonRepo) Create(ctx context.Context, l *domain.Lesson) error {
	return m.Called(ctx, l).Error(0)
}
func (m *MockLessonRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Lesson, error) {
	args := m.Called(ctx, id)
	if l := args.Get(0); l != nil {
		return l.(*domain.Lesson), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockLessonRepo) Update(ctx context.Context, l *domain.Lesson) error {
	return m.Called(ctx, l).Error(0)
}
func (m *MockLessonRepo) HasClassroomCollision(ctx context.Context, classroomID, teacherID uuid.UUID, startTime, endTime time.Time, excludeLessonID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, classroomID, teacherID, startTime, endTime, excludeLessonID)
	return args.Bool(0), args.Error(1)
}
func (m *MockLessonRepo) List(ctx context.Context, filter schedule.LessonFilter) ([]*domain.Lesson, error) {
	args := m.Called(ctx, filter)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Lesson), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestScheduleHandler_Classrooms(t *testing.T) {
	cRepo := new(MockClassroomRepo)
	tokenMgr := new(MockTokenManager)

	ownerID := uuid.New()
	ownerClaims := &security.UserClaims{
		UserID: ownerID,
		Role:   domain.RoleOwner,
	}

	tokenMgr.On("ValidateAccessToken", "owner_token").Return(ownerClaims, nil)

	scheduleSvc := schedule.NewService(cRepo, nil, nil, nil)
	handler := httpserver.NewAPIHandler(nil, scheduleSvc, nil, nil, tokenMgr, "v1")

	t.Run("Create classroom by owner success", func(t *testing.T) {
		cRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Classroom")).Return(nil).Once()

		reqBody := `{"name":"Физика 101","capacity":4,"color":"#10B981"}`
		req := httptest.NewRequest(http.MethodPost, "/classrooms", bytes.NewBufferString(reqBody))
		req.Header.Set("Authorization", "Bearer owner_token")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.CreateClassroom(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var resp generated.ClassroomResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Физика 101", resp.Name)
		assert.Equal(t, 4, resp.Capacity)
	})

	t.Run("List classrooms", func(t *testing.T) {
		classrooms := []*domain.Classroom{
			{ID: uuid.New(), Name: "Кабинет 1", Capacity: 2, Color: "#3B82F6"},
		}
		cRepo.On("List", mock.Anything).Return(classrooms, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/classrooms", nil)
		req.Header.Set("Authorization", "Bearer owner_token")
		rec := httptest.NewRecorder()

		handler.ListClassrooms(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp []generated.ClassroomResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "Кабинет 1", resp[0].Name)
	})
}

func TestScheduleHandler_Lessons(t *testing.T) {
	cRepo := new(MockClassroomRepo)
	clientRepo := new(MockClientRepo)
	lRepo := new(MockLessonRepo)
	tokenMgr := new(MockTokenManager)

	teacherID := uuid.New()
	clientID := uuid.New()
	classroomID := uuid.New()

	teacherClaims := &security.UserClaims{UserID: teacherID, Role: domain.RoleTeacher}

	tokenMgr.On("ValidateAccessToken", "teacher_token").Return(teacherClaims, nil)

	scheduleSvc := schedule.NewService(cRepo, lRepo, clientRepo, nil)
	handler := httpserver.NewAPIHandler(nil, scheduleSvc, nil, nil, tokenMgr, "v1")

	start := time.Now().UTC().Add(time.Hour)
	end := start.Add(time.Hour)

	t.Run("Create lesson success", func(t *testing.T) {
		clientRepo.On("GetByID", mock.Anything, clientID).Return(&domain.Client{
			ID:        clientID,
			TeacherID: teacherID,
		}, nil).Once()
		cRepo.On("GetByID", mock.Anything, classroomID).Return(&domain.Classroom{ID: classroomID, Name: "Кабинет"}, nil).Once()
		lRepo.On("HasClassroomCollision", mock.Anything, classroomID, teacherID, start, end, (*uuid.UUID)(nil)).Return(false, nil).Once()
		lRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Lesson")).Return(nil).Once()

		reqObj := generated.CreateLessonRequest{
			ClientId:    clientID,
			ClassroomId: &classroomID,
			Format:      generated.Offline,
			StartTime:   start,
			EndTime:     end,
		}
		bodyBytes, _ := json.Marshal(reqObj)
		req := httptest.NewRequest(http.MethodPost, "/lessons", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer teacher_token")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.CreateLesson(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var resp generated.LessonResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, generated.Scheduled, resp.Status)
	})

	t.Run("Create lesson collision returns 409", func(t *testing.T) {
		clientRepo.On("GetByID", mock.Anything, clientID).Return(&domain.Client{
			ID:        clientID,
			TeacherID: teacherID,
		}, nil).Once()
		cRepo.On("GetByID", mock.Anything, classroomID).Return(&domain.Classroom{ID: classroomID, Name: "Кабинет"}, nil).Once()
		lRepo.On("HasClassroomCollision", mock.Anything, classroomID, teacherID, start, end, (*uuid.UUID)(nil)).Return(true, nil).Once()

		reqObj := generated.CreateLessonRequest{
			ClientId:    clientID,
			ClassroomId: &classroomID,
			Format:      generated.Offline,
			StartTime:   start,
			EndTime:     end,
		}
		bodyBytes, _ := json.Marshal(reqObj)
		req := httptest.NewRequest(http.MethodPost, "/lessons", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer teacher_token")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.CreateLesson(rec, req)

		assert.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("Complete lesson by teacher", func(t *testing.T) {
		lessonID := uuid.New()
		l := &domain.Lesson{
			ID:        lessonID,
			TeacherID: teacherID,
			ClientID:  clientID,
			Status:    domain.StatusScheduled,
		}
		lRepo.On("GetByID", mock.Anything, lessonID).Return(l, nil).Once()
		lRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Lesson")).Return(nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/lessons/"+lessonID.String()+"/complete", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rec := httptest.NewRecorder()

		handler.CompleteLesson(rec, req, lessonID)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp generated.LessonResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, generated.Completed, resp.Status)
	})
}
