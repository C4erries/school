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

// MockTeacherStudentRepository
type MockTeacherStudentRepository struct {
	mock.Mock
}

func (m *MockTeacherStudentRepository) Create(ctx context.Context, ts *domain.TeacherStudent) error {
	return m.Called(ctx, ts).Error(0)
}

func (m *MockTeacherStudentRepository) Delete(ctx context.Context, teacherID, studentID uuid.UUID) error {
	return m.Called(ctx, teacherID, studentID).Error(0)
}

func (m *MockTeacherStudentRepository) IsAssigned(ctx context.Context, teacherID, studentID uuid.UUID) (bool, error) {
	args := m.Called(ctx, teacherID, studentID)
	return args.Bool(0), args.Error(1)
}

func (m *MockTeacherStudentRepository) ListStudentsByTeacher(ctx context.Context, teacherID uuid.UUID) ([]*domain.User, error) {
	args := m.Called(ctx, teacherID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTeacherStudentRepository) ListTeachersByStudent(ctx context.Context, studentID uuid.UUID) ([]*domain.User, error) {
	args := m.Called(ctx, studentID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
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

// MockUserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if u := args.Get(0); u != nil {
		return u.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestScheduleService_Classrooms(t *testing.T) {
	ctx := context.Background()

	t.Run("Create classroom success", func(t *testing.T) {
		cRepo := new(MockClassroomRepository)
		tsRepo := new(MockTeacherStudentRepository)
		lRepo := new(MockLessonRepository)
		uRepo := new(MockUserRepository)

		cRepo.On("Create", ctx, mock.MatchedBy(func(c *domain.Classroom) bool {
			return c.Name == "Кабинет 101" && c.Capacity == 5 && c.Color == "#3B82F6"
		})).Return(nil)

		svc := schedule.NewService(cRepo, tsRepo, lRepo, uRepo)
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

func TestScheduleService_Assignments(t *testing.T) {
	ctx := context.Background()
	teacherID := uuid.New()
	studentID := uuid.New()

	t.Run("Assign student success", func(t *testing.T) {
		tsRepo := new(MockTeacherStudentRepository)
		uRepo := new(MockUserRepository)

		uRepo.On("GetByID", ctx, teacherID).Return(&domain.User{ID: teacherID, Role: domain.RoleTeacher}, nil)
		uRepo.On("GetByID", ctx, studentID).Return(&domain.User{ID: studentID, Role: domain.RoleStudent}, nil)
		tsRepo.On("Create", ctx, mock.MatchedBy(func(ts *domain.TeacherStudent) bool {
			return ts.TeacherID == teacherID && ts.StudentID == studentID
		})).Return(nil)

		svc := schedule.NewService(nil, tsRepo, nil, uRepo)
		ts, err := svc.AssignStudent(ctx, teacherID, studentID)
		require.NoError(t, err)
		assert.Equal(t, teacherID, ts.TeacherID)
		assert.Equal(t, studentID, ts.StudentID)
		tsRepo.AssertExpectations(t)
		uRepo.AssertExpectations(t)
	})

	t.Run("Cannot assign self", func(t *testing.T) {
		svc := schedule.NewService(nil, nil, nil, nil)
		_, err := svc.AssignStudent(ctx, teacherID, teacherID)
		assert.ErrorIs(t, err, domain.ErrCannotAssignSelf)
	})

	t.Run("Invalid teacher role", func(t *testing.T) {
		uRepo := new(MockUserRepository)
		uRepo.On("GetByID", ctx, teacherID).Return(&domain.User{ID: teacherID, Role: domain.RoleStudent}, nil)

		svc := schedule.NewService(nil, nil, nil, uRepo)
		_, err := svc.AssignStudent(ctx, teacherID, studentID)
		assert.ErrorIs(t, err, domain.ErrInvalidTeacherRole)
		uRepo.AssertExpectations(t)
	})

	t.Run("Invalid student role", func(t *testing.T) {
		uRepo := new(MockUserRepository)
		uRepo.On("GetByID", ctx, teacherID).Return(&domain.User{ID: teacherID, Role: domain.RoleTeacher}, nil)
		uRepo.On("GetByID", ctx, studentID).Return(&domain.User{ID: studentID, Role: domain.RoleTeacher}, nil)

		svc := schedule.NewService(nil, nil, nil, uRepo)
		_, err := svc.AssignStudent(ctx, teacherID, studentID)
		assert.ErrorIs(t, err, domain.ErrInvalidStudentRole)
		uRepo.AssertExpectations(t)
	})

	t.Run("Already assigned", func(t *testing.T) {
		tsRepo := new(MockTeacherStudentRepository)
		uRepo := new(MockUserRepository)

		uRepo.On("GetByID", ctx, teacherID).Return(&domain.User{ID: teacherID, Role: domain.RoleTeacher}, nil)
		uRepo.On("GetByID", ctx, studentID).Return(&domain.User{ID: studentID, Role: domain.RoleStudent}, nil)
		tsRepo.On("Create", ctx, mock.Anything).Return(domain.ErrTeacherStudentAlreadyExists)

		svc := schedule.NewService(nil, tsRepo, nil, uRepo)
		_, err := svc.AssignStudent(ctx, teacherID, studentID)
		assert.ErrorIs(t, err, domain.ErrTeacherStudentAlreadyExists)
	})

	t.Run("Unassign and list students", func(t *testing.T) {
		tsRepo := new(MockTeacherStudentRepository)
		tsRepo.On("Delete", ctx, teacherID, studentID).Return(nil)
		tsRepo.On("ListStudentsByTeacher", ctx, teacherID).Return([]*domain.User{
			{ID: studentID, FullName: "Ученик"},
		}, nil)

		svc := schedule.NewService(nil, tsRepo, nil, nil)
		err := svc.UnassignStudent(ctx, teacherID, studentID)
		require.NoError(t, err)

		students, err := svc.ListTeacherStudents(ctx, teacherID)
		require.NoError(t, err)
		assert.Len(t, students, 1)
		tsRepo.AssertExpectations(t)
	})
}

func TestScheduleService_Lessons(t *testing.T) {
	ctx := context.Background()
	teacherID := uuid.New()
	studentID := uuid.New()
	classroomID := uuid.New()

	now := time.Now().UTC()
	start := now.Add(2 * time.Hour)
	end := start.Add(time.Hour)

	t.Run("Schedule online lesson success", func(t *testing.T) {
		tsRepo := new(MockTeacherStudentRepository)
		lRepo := new(MockLessonRepository)

		tsRepo.On("IsAssigned", ctx, teacherID, studentID).Return(true, nil)
		lRepo.On("Create", ctx, mock.MatchedBy(func(l *domain.Lesson) bool {
			return l.TeacherID == teacherID && l.StudentID == studentID && l.ClassroomID == nil && l.Format == domain.FormatOnline
		})).Return(nil)

		svc := schedule.NewService(nil, tsRepo, lRepo, nil)
		lesson, err := svc.ScheduleLesson(ctx, schedule.ScheduleLessonInput{
			TeacherID:     teacherID,
			StudentID:     studentID,
			StartTime:     start,
			EndTime:       end,
			Format:        domain.FormatOnline,
			LocationOrURL: "https://meet.google.com/abc",
		})

		require.NoError(t, err)
		assert.Equal(t, domain.FormatOnline, lesson.Format)
		assert.Equal(t, domain.StatusPendingConfirmation, lesson.Status)
		assert.Nil(t, lesson.ClassroomID)
		lRepo.AssertExpectations(t)
		tsRepo.AssertExpectations(t)
	})

	t.Run("Schedule offline lesson success without collision", func(t *testing.T) {
		cRepo := new(MockClassroomRepository)
		tsRepo := new(MockTeacherStudentRepository)
		lRepo := new(MockLessonRepository)

		tsRepo.On("IsAssigned", ctx, teacherID, studentID).Return(true, nil)
		cRepo.On("GetByID", ctx, classroomID).Return(&domain.Classroom{ID: classroomID, Name: "Кабинет 1"}, nil)
		lRepo.On("HasClassroomCollision", ctx, classroomID, teacherID, start, end, (*uuid.UUID)(nil)).Return(false, nil)
		lRepo.On("Create", ctx, mock.MatchedBy(func(l *domain.Lesson) bool {
			return l.ClassroomID != nil && *l.ClassroomID == classroomID && l.Format == domain.FormatOffline
		})).Return(nil)

		svc := schedule.NewService(cRepo, tsRepo, lRepo, nil)
		lesson, err := svc.ScheduleLesson(ctx, schedule.ScheduleLessonInput{
			TeacherID:   teacherID,
			StudentID:   studentID,
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
		tsRepo.AssertExpectations(t)
	})

	t.Run("Schedule offline lesson blocked by collision with another teacher", func(t *testing.T) {
		cRepo := new(MockClassroomRepository)
		tsRepo := new(MockTeacherStudentRepository)
		lRepo := new(MockLessonRepository)

		tsRepo.On("IsAssigned", ctx, teacherID, studentID).Return(true, nil)
		cRepo.On("GetByID", ctx, classroomID).Return(&domain.Classroom{ID: classroomID, Name: "Кабинет 1"}, nil)
		lRepo.On("HasClassroomCollision", ctx, classroomID, teacherID, start, end, (*uuid.UUID)(nil)).Return(true, nil)

		svc := schedule.NewService(cRepo, tsRepo, lRepo, nil)
		_, err := svc.ScheduleLesson(ctx, schedule.ScheduleLessonInput{
			TeacherID:   teacherID,
			StudentID:   studentID,
			ClassroomID: &classroomID,
			StartTime:   start,
			EndTime:     end,
			Format:      domain.FormatOffline,
		})

		assert.ErrorIs(t, err, domain.ErrClassroomCollision)
		lRepo.AssertExpectations(t)
	})

	t.Run("Schedule lesson fails if student is not assigned to teacher", func(t *testing.T) {
		tsRepo := new(MockTeacherStudentRepository)
		tsRepo.On("IsAssigned", ctx, teacherID, studentID).Return(false, nil)

		svc := schedule.NewService(nil, tsRepo, nil, nil)
		_, err := svc.ScheduleLesson(ctx, schedule.ScheduleLessonInput{
			TeacherID: teacherID,
			StudentID: studentID,
			StartTime: start,
			EndTime:   end,
			Format:    domain.FormatOnline,
		})

		assert.ErrorIs(t, err, domain.ErrStudentNotAssignedToTeacher)
		tsRepo.AssertExpectations(t)
	})

	t.Run("Accept and Decline lesson", func(t *testing.T) {
		lRepo := new(MockLessonRepository)
		lessonID := uuid.New()
		pendingLesson := &domain.Lesson{
			ID:        lessonID,
			TeacherID: teacherID,
			StudentID: studentID,
			Status:    domain.StatusPendingConfirmation,
		}

		// Accept by student
		lRepo.On("GetByID", ctx, lessonID).Return(pendingLesson, nil).Once()
		lRepo.On("Update", ctx, mock.MatchedBy(func(l *domain.Lesson) bool {
			return l.Status == domain.StatusConfirmed
		})).Return(nil).Once()

		svc := schedule.NewService(nil, nil, lRepo, nil)
		accepted, err := svc.AcceptLesson(ctx, lessonID, studentID)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusConfirmed, accepted.Status)

		// Decline by stranger fails
		pendingLesson.Status = domain.StatusPendingConfirmation
		lRepo.On("GetByID", ctx, lessonID).Return(pendingLesson, nil).Once()
		strangerID := uuid.New()
		_, err = svc.DeclineLesson(ctx, lessonID, strangerID, "Не могу")
		assert.ErrorIs(t, err, domain.ErrUnauthorizedLessonAction)

		// Decline by student succeeds
		lRepo.On("GetByID", ctx, lessonID).Return(pendingLesson, nil).Once()
		lRepo.On("Update", ctx, mock.MatchedBy(func(l *domain.Lesson) bool {
			return l.Status == domain.StatusDeclined && l.CancelReason == "Заболел"
		})).Return(nil).Once()

		declined, err := svc.DeclineLesson(ctx, lessonID, studentID, "Заболел")
		require.NoError(t, err)
		assert.Equal(t, domain.StatusDeclined, declined.Status)
		lRepo.AssertExpectations(t)
	})

	t.Run("Complete lesson", func(t *testing.T) {
		lRepo := new(MockLessonRepository)
		lessonID := uuid.New()
		confirmedLesson := &domain.Lesson{
			ID:        lessonID,
			TeacherID: teacherID,
			StudentID: studentID,
			Status:    domain.StatusConfirmed,
		}

		lRepo.On("GetByID", ctx, lessonID).Return(confirmedLesson, nil).Once()
		lRepo.On("Update", ctx, mock.MatchedBy(func(l *domain.Lesson) bool {
			return l.Status == domain.StatusCompleted
		})).Return(nil).Once()

		svc := schedule.NewService(nil, nil, lRepo, nil)
		completed, err := svc.CompleteLesson(ctx, lessonID, teacherID, domain.RoleTeacher)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusCompleted, completed.Status)
		lRepo.AssertExpectations(t)
	})

	t.Run("Cancel lesson by student and teacher", func(t *testing.T) {
		lRepo := new(MockLessonRepository)
		lessonID := uuid.New()

		// Cancel by teacher
		l := &domain.Lesson{
			ID:        lessonID,
			TeacherID: teacherID,
			StudentID: studentID,
			Status:    domain.StatusConfirmed,
		}
		lRepo.On("GetByID", ctx, lessonID).Return(l, nil).Once()
		lRepo.On("Update", ctx, mock.MatchedBy(func(l *domain.Lesson) bool {
			return l.Status == domain.StatusCancelledByTeacher
		})).Return(nil).Once()

		svc := schedule.NewService(nil, nil, lRepo, nil)
		res, err := svc.CancelLesson(ctx, lessonID, teacherID, domain.RoleTeacher, "Преподаватель отменил")
		require.NoError(t, err)
		assert.Equal(t, domain.StatusCancelledByTeacher, res.Status)

		// Cancel by student
		l2 := &domain.Lesson{
			ID:        lessonID,
			TeacherID: teacherID,
			StudentID: studentID,
			Status:    domain.StatusPendingConfirmation,
		}
		lRepo.On("GetByID", ctx, lessonID).Return(l2, nil).Once()
		lRepo.On("Update", ctx, mock.MatchedBy(func(l *domain.Lesson) bool {
			return l.Status == domain.StatusCancelledByStudent
		})).Return(nil).Once()

		res2, err := svc.CancelLesson(ctx, lessonID, studentID, domain.RoleStudent, "Ученик отменил")
		require.NoError(t, err)
		assert.Equal(t, domain.StatusCancelledByStudent, res2.Status)

		lRepo.AssertExpectations(t)
	})

	t.Run("List lessons with filter", func(t *testing.T) {
		lRepo := new(MockLessonRepository)
		filter := schedule.LessonFilter{TeacherID: &teacherID}
		expected := []*domain.Lesson{
			{ID: uuid.New(), TeacherID: teacherID},
		}
		lRepo.On("List", ctx, filter).Return(expected, nil)

		svc := schedule.NewService(nil, nil, lRepo, nil)
		res, err := svc.ListLessons(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, expected, res)
		lRepo.AssertExpectations(t)
	})
}
