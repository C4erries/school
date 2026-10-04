package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/domain"
)

func TestLesson_StateMachine(t *testing.T) {
	now := time.Now().UTC()
	start := now.Add(2 * time.Hour)
	end := start.Add(time.Hour)

	createPendingLesson := func() *domain.Lesson {
		return &domain.Lesson{
			ID:          uuid.New(),
			TeacherID:   uuid.New(),
			StudentID:   uuid.New(),
			StartTime:   start,
			EndTime:     end,
			Format:      domain.FormatOffline,
			Status:      domain.StatusPendingConfirmation,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}

	t.Run("Pending -> Confirmed -> Completed", func(t *testing.T) {
		lesson := createPendingLesson()
		require.True(t, lesson.CanAccept())
		require.NoError(t, lesson.Accept())
		assert.Equal(t, domain.StatusConfirmed, lesson.Status)

		require.True(t, lesson.CanComplete())
		require.NoError(t, lesson.Complete())
		assert.Equal(t, domain.StatusCompleted, lesson.Status)

		// Cannot complete or accept again
		assert.False(t, lesson.CanAccept())
		assert.False(t, lesson.CanComplete())
		assert.ErrorIs(t, lesson.Accept(), domain.ErrInvalidLessonStatus)
		assert.ErrorIs(t, lesson.Complete(), domain.ErrInvalidLessonStatus)
	})

	t.Run("Pending -> Declined", func(t *testing.T) {
		lesson := createPendingLesson()
		require.True(t, lesson.CanDecline())
		require.NoError(t, lesson.Decline("Не успеваю"))
		assert.Equal(t, domain.StatusDeclined, lesson.Status)
		assert.Equal(t, "Не успеваю", lesson.CancelReason)

		// Cannot transition from declined
		assert.False(t, lesson.CanAccept())
		assert.ErrorIs(t, lesson.Accept(), domain.ErrInvalidLessonStatus)
	})

	t.Run("Pending -> Cancelled by teacher", func(t *testing.T) {
		lesson := createPendingLesson()
		require.True(t, lesson.CanCancel())
		require.NoError(t, lesson.Cancel(domain.RoleTeacher, "Заболел"))
		assert.Equal(t, domain.StatusCancelledByTeacher, lesson.Status)
		assert.Equal(t, "Заболел", lesson.CancelReason)
	})

	t.Run("Confirmed -> Cancelled by student", func(t *testing.T) {
		lesson := createPendingLesson()
		require.NoError(t, lesson.Accept())
		require.True(t, lesson.CanCancel())
		require.NoError(t, lesson.Cancel(domain.RoleStudent, "Срочные дела"))
		assert.Equal(t, domain.StatusCancelledByStudent, lesson.Status)
		assert.Equal(t, "Срочные дела", lesson.CancelReason)
	})

	t.Run("Confirmed -> NoShow", func(t *testing.T) {
		lesson := createPendingLesson()
		require.NoError(t, lesson.Accept())
		require.True(t, lesson.CanMarkNoShow())
		require.NoError(t, lesson.MarkNoShow("Ученик не подключился"))
		assert.Equal(t, domain.StatusNoShow, lesson.Status)
		assert.Equal(t, "Ученик не подключился", lesson.CancelReason)
	})

	t.Run("ValidateLessonTimes", func(t *testing.T) {
		assert.NoError(t, domain.ValidateLessonTimes(start, end))
		assert.ErrorIs(t, domain.ValidateLessonTimes(end, start), domain.ErrInvalidTimeRange)
		assert.ErrorIs(t, domain.ValidateLessonTimes(start, start), domain.ErrInvalidTimeRange)
	})
}
