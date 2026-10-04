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

// MockClassroomRepo
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

// MockTeacherStudentRepo
type MockTeacherStudentRepo struct {
	mock.Mock
}

func (m *MockTeacherStudentRepo) Create(ctx context.Context, ts *domain.TeacherStudent) error {
	return m.Called(ctx, ts).Error(0)
}
func (m *MockTeacherStudentRepo) Delete(ctx context.Context, teacherID, studentID uuid.UUID) error {
	return m.Called(ctx, teacherID, studentID).Error(0)
}
func (m *MockTeacherStudentRepo) IsAssigned(ctx context.Context, teacherID, studentID uuid.UUID) (bool, error) {
	args := m.Called(ctx, teacherID, studentID)
	return args.Bool(0), args.Error(1)
}
func (m *MockTeacherStudentRepo) ListStudentsByTeacher(ctx context.Context, teacherID uuid.UUID) ([]*domain.User, error) {
	args := m.Called(ctx, teacherID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockTeacherStudentRepo) ListTeachersByStudent(ctx context.Context, studentID uuid.UUID) ([]*domain.User, error) {
	args := m.Called(ctx, studentID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

// MockLessonRepo
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
	uRepo := new(MockUserRepository)
	tokenMgr := new(MockTokenManager)

	ownerID := uuid.New()
	ownerClaims := &security.UserClaims{
		UserID: ownerID,
		Role:   domain.RoleOwner,
	}

	tokenMgr.On("ValidateAccessToken", "owner_token").Return(ownerClaims, nil)

	scheduleSvc := schedule.NewService(cRepo, nil, nil, uRepo)
	handler := httpserver.NewAPIHandler(nil, scheduleSvc, tokenMgr, "v1")

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

	t.Run("Create classroom forbidden for student", func(t *testing.T) {
		studentClaims := &security.UserClaims{
			UserID: uuid.New(),
			Role:   domain.RoleStudent,
		}
		tokenMgr.On("ValidateAccessToken", "student_token").Return(studentClaims, nil).Once()

		reqBody := `{"name":"Физика 101","capacity":4}`
		req := httptest.NewRequest(http.MethodPost, "/classrooms", bytes.NewBufferString(reqBody))
		req.Header.Set("Authorization", "Bearer student_token")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.CreateClassroom(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("List classrooms", func(t *testing.T) {
		cid := uuid.New()
		cRepo.On("List", mock.Anything).Return([]*domain.Classroom{
			{ID: cid, Name: "Кабинет 1", Capacity: 2, Color: "#3B82F6", CreatedAt: time.Now().UTC()},
		}, nil).Once()

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
	tsRepo := new(MockTeacherStudentRepo)
	lRepo := new(MockLessonRepo)
	uRepo := new(MockUserRepository)
	tokenMgr := new(MockTokenManager)

	teacherID := uuid.New()
	studentID := uuid.New()
	classroomID := uuid.New()

	teacherClaims := &security.UserClaims{UserID: teacherID, Role: domain.RoleTeacher}
	studentClaims := &security.UserClaims{UserID: studentID, Role: domain.RoleStudent}

	tokenMgr.On("ValidateAccessToken", "teacher_token").Return(teacherClaims, nil)
	tokenMgr.On("ValidateAccessToken", "student_token").Return(studentClaims, nil)

	scheduleSvc := schedule.NewService(cRepo, tsRepo, lRepo, uRepo)
	handler := httpserver.NewAPIHandler(nil, scheduleSvc, tokenMgr, "v1")

	start := time.Now().UTC().Add(time.Hour)
	end := start.Add(time.Hour)

	t.Run("Create lesson success", func(t *testing.T) {
		tsRepo.On("IsAssigned", mock.Anything, teacherID, studentID).Return(true, nil).Once()
		cRepo.On("GetByID", mock.Anything, classroomID).Return(&domain.Classroom{ID: classroomID, Name: "Кабинет"}, nil).Once()
		lRepo.On("HasClassroomCollision", mock.Anything, classroomID, teacherID, start, end, (*uuid.UUID)(nil)).Return(false, nil).Once()
		lRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Lesson")).Return(nil).Once()

		reqObj := generated.CreateLessonRequest{
			StudentId:   studentID,
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
		assert.Equal(t, generated.PendingConfirmation, resp.Status)
	})

	t.Run("Create lesson collision returns 409", func(t *testing.T) {
		tsRepo.On("IsAssigned", mock.Anything, teacherID, studentID).Return(true, nil).Once()
		cRepo.On("GetByID", mock.Anything, classroomID).Return(&domain.Classroom{ID: classroomID, Name: "Кабинет"}, nil).Once()
		lRepo.On("HasClassroomCollision", mock.Anything, classroomID, teacherID, start, end, (*uuid.UUID)(nil)).Return(true, nil).Once()

		reqObj := generated.CreateLessonRequest{
			StudentId:   studentID,
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

	t.Run("Accept lesson by student", func(t *testing.T) {
		lessonID := uuid.New()
		l := &domain.Lesson{
			ID:        lessonID,
			TeacherID: teacherID,
			StudentID: studentID,
			Status:    domain.StatusPendingConfirmation,
		}
		lRepo.On("GetByID", mock.Anything, lessonID).Return(l, nil).Once()
		lRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Lesson")).Return(nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/lessons/"+lessonID.String()+"/accept", nil)
		req.Header.Set("Authorization", "Bearer student_token")
		rec := httptest.NewRecorder()

		handler.AcceptLesson(rec, req, lessonID)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp generated.LessonResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, generated.Confirmed, resp.Status)
	})
}
