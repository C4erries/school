package schedule_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
)

// MockClassroomRepository
type MockClassroomRepository struct {
	mock.Mock
}

func (m *MockClassroomRepository) Create(ctx context.Context, c *domain.Classroom) error {
	return m.Called(ctx, c).Error(0)
}

func (m *MockClassroomRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Classroom, error) {
	args := m.Called(ctx, id)
	if c := args.Get(0); c != nil {
		return c.(*domain.Classroom), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockClassroomRepository) List(ctx context.Context) ([]*domain.Classroom, error) {
	args := m.Called(ctx)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Classroom), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockClassroomRepository) Update(ctx context.Context, c *domain.Classroom) error {
	return m.Called(ctx, c).Error(0)
}

func (m *MockClassroomRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// MockClientRepository
type MockClientRepository struct {
	mock.Mock
}

func (m *MockClientRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	args := m.Called(ctx, id)
	if c := args.Get(0); c != nil {
		return c.(*domain.Client), args.Error(1)
	}
	return nil, args.Error(1)
}

// MockSubscriptionRepository
type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) GetByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.ClientSubscription, error) {
	args := m.Called(ctx, clientID)
	if subs := args.Get(0); subs != nil {
		return subs.([]*domain.ClientSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSubscriptionRepository) Update(ctx context.Context, sub *domain.ClientSubscription) error {
	return m.Called(ctx, sub).Error(0)
}

// MockLessonRepository
type MockLessonRepository struct {
	mock.Mock
}

func (m *MockLessonRepository) Create(ctx context.Context, l *domain.Lesson) error {
	return m.Called(ctx, l).Error(0)
}

func (m *MockLessonRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Lesson, error) {
	args := m.Called(ctx, id)
	if l := args.Get(0); l != nil {
		return l.(*domain.Lesson), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockLessonRepository) Update(ctx context.Context, l *domain.Lesson) error {
	return m.Called(ctx, l).Error(0)
}

func (m *MockLessonRepository) HasClassroomCollision(ctx context.Context, classroomID, teacherID uuid.UUID, startTime, endTime time.Time, excludeLessonID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, classroomID, teacherID, startTime, endTime, excludeLessonID)
	return args.Bool(0), args.Error(1)
}

func (m *MockLessonRepository) List(ctx context.Context, filter schedule.LessonFilter) ([]*domain.Lesson, error) {
	args := m.Called(ctx, filter)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Lesson), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestScheduleService_Classrooms(t *testing.T) {
	ctx := context.Background()

	t.Run("Create classroom success", func(t *testing.T) {
		cRepo := new(MockClassroomRepository)

		cRepo.On("Create", ctx, mock.MatchedBy(func(c *domain.Classroom) bool {
			return c.Name == "Кабинет 101" && c.Capacity == 5 && c.Color == "#3B82F6"
		})).Return(nil)

		svc := schedule.NewService(cRepo, nil, nil, nil)
		c, err := svc.CreateClassroom(ctx, schedule.CreateClassroomInput{
			Name:     "Кабинет 101",
			Capacity: 5,
		})

		require.NoError(t, err)
		assert.Equal(t, "Кабинет 101", c.Name)
		assert.Equal(t, 5, c.Capacity)
		assert.Equal(t, "#3B82F6", c.Color)
		cRepo.AssertExpectations(t)
	})

	t.Run("Create classroom invalid name", func(t *testing.T) {
		svc := schedule.NewService(nil, nil, nil, nil)
		_, err := svc.CreateClassroom(ctx, schedule.CreateClassroomInput{
			Name:     "A",
			Capacity: 5,
		})
		assert.ErrorIs(t, err, domain.ErrInvalidClassroomName)
	})

	t.Run("Create classroom invalid capacity", func(t *testing.T) {
		svc := schedule.NewService(nil, nil, nil, nil)
		_, err := svc.CreateClassroom(ctx, schedule.CreateClassroomInput{
			Name:     "Класс",
			Capacity: 0,
		})
		assert.ErrorIs(t, err, domain.ErrInvalidClassroomCapacity)
	})

	t.Run("Get and List classrooms", func(t *testing.T) {
		cRepo := new(MockClassroomRepository)
		cid := uuid.New()
		classrooms := []*domain.Classroom{
			{ID: cid, Name: "Кабинет 1"},
		}

		cRepo.On("GetByID", ctx, cid).Return(classrooms[0], nil)
		cRepo.On("List", ctx).Return(classrooms, nil)

		svc := schedule.NewService(cRepo, nil, nil, nil)
		item, err := svc.GetClassroom(ctx, cid)
		require.NoError(t, err)
		assert.Equal(t, cid, item.ID)

		list, err := svc.ListClassrooms(ctx)
		require.NoError(t, err)
		assert.Len(t, list, 1)
		cRepo.AssertExpectations(t)
	})

	t.Run("Delete classroom", func(t *testing.T) {
		cRepo := new(MockClassroomRepository)
		cid := uuid.New()
		cRepo.On("Delete", ctx, cid).Return(nil)

		svc := schedule.NewService(cRepo, nil, nil, nil)
		err := svc.DeleteClassroom(ctx, cid)
		require.NoError(t, err)
		cRepo.AssertExpectations(t)
	})
}

func TestScheduleService_Lessons(t *testing.T) {
	ctx := context.Background()
	teacherID := uuid.New()
	clientID := uuid.New()
	classroomID := uuid.New()

	now := time.Now().UTC()
	start := now.Add(2 * time.Hour)
	end := start.Add(time.Hour)

	t.Run("Schedule online lesson success", func(t *testing.T) {
		clientRepo := new(MockClientRepository)
		lRepo := new(MockLessonRepository)

		clientRepo.On("GetByID", ctx, clientID).Return(&domain.Client{
			ID:        clientID,
			TeacherID: teacherID,
		}, nil)
		lRepo.On("Create", ctx, mock.MatchedBy(func(l *domain.Lesson) bool {
			return l.TeacherID == teacherID && l.ClientID == clientID && l.ClassroomID == nil && l.Format == domain.FormatOnline
		})).Return(nil)

		svc := schedule.NewService(nil, lRepo, clientRepo, nil)
		lesson, err := svc.ScheduleLesson(ctx, schedule.ScheduleLessonInput{
			TeacherID:     teacherID,
			ClientID:      clientID,
			StartTime:     start,
			EndTime:       end,
			Format:        domain.FormatOnline,
			LocationOrURL: "https://meet.google.com/abc",
		})

		require.NoError(t, err)
		assert.Equal(t, domain.FormatOnline, lesson.Format)
		assert.Equal(t, domain.StatusScheduled, lesson.Status)
		assert.Nil(t, lesson.ClassroomID)
		lRepo.AssertExpectations(t)
		clientRepo.AssertExpectations(t)
	})

	t.Run("Schedule offline lesson success without collision", func(t *testing.T) {
		cRepo := new(MockClassroomRepository)
		clientRepo := new(MockClientRepository)
		lRepo := new(MockLessonRepository)

		clientRepo.On("GetByID", ctx, clientID).Return(&domain.Client{
			ID:        clientID,
			TeacherID: teacherID,
		}, nil)
		cRepo.On("GetByID", ctx, classroomID).Return(&domain.Classroom{ID: classroomID, Name: "Кабинет 1"}, nil)
		lRepo.On("HasClassroomCollision", ctx, classroomID, teacherID, start, end, (*uuid.UUID)(nil)).Return(false, nil)
		lRepo.On("Create", ctx, mock.MatchedBy(func(l *domain.Lesson) bool {
			return l.ClassroomID != nil && *l.ClassroomID == classroomID && l.Format == domain.FormatOffline
		})).Return(nil)

		svc := schedule.NewService(cRepo, lRepo, clientRepo, nil)
		lesson, err := svc.ScheduleLesson(ctx, schedule.ScheduleLessonInput{
			TeacherID:   teacherID,
			ClientID:    clientID,
			ClassroomID: &classroomID,
			StartTime:   start,
			EndTime:     end,
			Format:      domain.FormatOffline,
		})

		require.NoError(t, err)
		assert.Equal(t, domain.FormatOffline, lesson.Format)
		assert.Equal(t, &classroomID, lesson.ClassroomID)
		cRepo.AssertExpectations(t)
		lRepo.AssertExpectations(t)
		clientRepo.AssertExpectations(t)
	})

	t.Run("Schedule offline lesson blocked by collision with another teacher", func(t *testing.T) {
		cRepo := new(MockClassroomRepository)
		clientRepo := new(MockClientRepository)
		lRepo := new(MockLessonRepository)

		clientRepo.On("GetByID", ctx, clientID).Return(&domain.Client{
			ID:        clientID,
			TeacherID: teacherID,
		}, nil)
		cRepo.On("GetByID", ctx, classroomID).Return(&domain.Classroom{ID: classroomID, Name: "Кабинет 1"}, nil)
		lRepo.On("HasClassroomCollision", ctx, classroomID, teacherID, start, end, (*uuid.UUID)(nil)).Return(true, nil)

		svc := schedule.NewService(cRepo, lRepo, clientRepo, nil)
		_, err := svc.ScheduleLesson(ctx, schedule.ScheduleLessonInput{
			TeacherID:   teacherID,
			ClientID:    clientID,
			ClassroomID: &classroomID,
			StartTime:   start,
			EndTime:     end,
			Format:      domain.FormatOffline,
		})

		assert.ErrorIs(t, err, domain.ErrClassroomCollision)
		lRepo.AssertExpectations(t)
	})

	t.Run("Schedule lesson fails if client belongs to another teacher", func(t *testing.T) {
		clientRepo := new(MockClientRepository)
		otherTeacherID := uuid.New()
		clientRepo.On("GetByID", ctx, clientID).Return(&domain.Client{
			ID:        clientID,
			TeacherID: otherTeacherID,
		}, nil)

		svc := schedule.NewService(nil, nil, clientRepo, nil)
		_, err := svc.ScheduleLesson(ctx, schedule.ScheduleLessonInput{
			TeacherID: teacherID,
			ClientID:  clientID,
			StartTime: start,
			EndTime:   end,
			Format:    domain.FormatOnline,
		})

		assert.ErrorIs(t, err, domain.ErrUnauthorizedLessonAction)
		clientRepo.AssertExpectations(t)
	})

	t.Run("Complete lesson and deduct subscription", func(t *testing.T) {
		lRepo := new(MockLessonRepository)
		subRepo := new(MockSubscriptionRepository)
		lessonID := uuid.New()
		scheduledLesson := &domain.Lesson{
			ID:        lessonID,
			TeacherID: teacherID,
			ClientID:  clientID,
			StartTime: start,
			EndTime:   end,
			Status:    domain.StatusScheduled,
		}

		sub := &domain.ClientSubscription{
			ID:       uuid.New(),
			ClientID: clientID,
			Type:     domain.SubscriptionTypeLessons,
			Balance:  5,
		}

		lRepo.On("GetByID", ctx, lessonID).Return(scheduledLesson, nil).Once()
		subRepo.On("GetByClientID", ctx, clientID).Return([]*domain.ClientSubscription{sub}, nil).Once()
		subRepo.On("Update", ctx, mock.MatchedBy(func(s *domain.ClientSubscription) bool {
			return s.Balance == 4
		})).Return(nil).Once()
		lRepo.On("Update", ctx, mock.MatchedBy(func(l *domain.Lesson) bool {
			return l.Status == domain.StatusCompleted
		})).Return(nil).Once()

		svc := schedule.NewService(nil, lRepo, nil, subRepo)
		completed, err := svc.CompleteLesson(ctx, lessonID, teacherID, domain.RoleTeacher)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusCompleted, completed.Status)
		lRepo.AssertExpectations(t)
		subRepo.AssertExpectations(t)
	})

	t.Run("Cancel lesson by teacher", func(t *testing.T) {
		lRepo := new(MockLessonRepository)
		lessonID := uuid.New()

		l := &domain.Lesson{
			ID:        lessonID,
			TeacherID: teacherID,
			ClientID:  clientID,
			Status:    domain.StatusScheduled,
		}
		lRepo.On("GetByID", ctx, lessonID).Return(l, nil).Once()
		lRepo.On("Update", ctx, mock.MatchedBy(func(l *domain.Lesson) bool {
			return l.Status == domain.StatusCancelled && l.CancelReason == "Преподаватель отменил"
		})).Return(nil).Once()

		svc := schedule.NewService(nil, lRepo, nil, nil)
		res, err := svc.CancelLesson(ctx, lessonID, teacherID, domain.RoleTeacher, "Преподаватель отменил")
		require.NoError(t, err)
		assert.Equal(t, domain.StatusCancelled, res.Status)

		lRepo.AssertExpectations(t)
	})

	t.Run("List lessons with filter", func(t *testing.T) {
		lRepo := new(MockLessonRepository)
		filter := schedule.LessonFilter{TeacherID: &teacherID}
		expected := []*domain.Lesson{
			{ID: uuid.New(), TeacherID: teacherID},
		}
		lRepo.On("List", ctx, filter).Return(expected, nil)

		svc := schedule.NewService(nil, lRepo, nil, nil)
		res, err := svc.ListLessons(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, expected, res)
		lRepo.AssertExpectations(t)
	})
}
