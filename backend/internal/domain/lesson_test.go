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

	createScheduledLesson := func() *domain.Lesson {
		return &domain.Lesson{
			ID:        uuid.New(),
			TeacherID: uuid.New(),
			ClientID:  uuid.New(),
			StartTime: start,
			EndTime:   end,
			Format:    domain.FormatOffline,
			Status:    domain.StatusScheduled,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	t.Run("Scheduled -> Completed", func(t *testing.T) {
		lesson := createScheduledLesson()
		require.True(t, lesson.CanComplete())
		require.NoError(t, lesson.Complete())
		assert.Equal(t, domain.StatusCompleted, lesson.Status)

		// Cannot complete or cancel again
		assert.False(t, lesson.CanComplete())
		assert.False(t, lesson.CanCancel())
		assert.ErrorIs(t, lesson.Complete(), domain.ErrInvalidLessonStatus)
		assert.ErrorIs(t, lesson.Cancel("отмена"), domain.ErrInvalidLessonStatus)
	})

	t.Run("Scheduled -> Cancelled", func(t *testing.T) {
		lesson := createScheduledLesson()
		require.True(t, lesson.CanCancel())
		require.NoError(t, lesson.Cancel("Заболел"))
		assert.Equal(t, domain.StatusCancelled, lesson.Status)
		assert.Equal(t, "Заболел", lesson.CancelReason)

		// Cannot transition from cancelled
		assert.False(t, lesson.CanComplete())
		assert.False(t, lesson.CanCancel())
		assert.ErrorIs(t, lesson.Complete(), domain.ErrInvalidLessonStatus)
		assert.ErrorIs(t, lesson.Cancel("еще раз"), domain.ErrInvalidLessonStatus)
	})

	t.Run("ValidateLessonTimes", func(t *testing.T) {
		assert.NoError(t, domain.ValidateLessonTimes(start, end))
		assert.ErrorIs(t, domain.ValidateLessonTimes(end, start), domain.ErrInvalidTimeRange)
		assert.ErrorIs(t, domain.ValidateLessonTimes(start, start), domain.ErrInvalidTimeRange)
	})
}
