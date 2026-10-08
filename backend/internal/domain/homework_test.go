package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/domain"
)

func TestHomeworkAssignment_ValidationAndTransitions(t *testing.T) {
	clientID := uuid.New()
	teacherID := uuid.New()
	lessonID := uuid.New()
	dueDate := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour)

	t.Run("Valid creation with full fields", func(t *testing.T) {
		hw, err := domain.NewHomeworkAssignment(
			clientID,
			teacherID,
			&lessonID,
			"Номера 15-20",
			"Решить задачи на тригонометрию",
			&dueDate,
		)
		require.NoError(t, err)
		assert.Equal(t, clientID, hw.ClientID)
		assert.Equal(t, teacherID, hw.TeacherID)
		assert.Equal(t, &lessonID, hw.AssignedLessonID)
		assert.Equal(t, "Номера 15-20", hw.Title)
		assert.Equal(t, "Решить задачи на тригонометрию", hw.Description)
		assert.Equal(t, domain.StatusHomeworkAssigned, hw.Status)
		assert.Equal(t, &dueDate, hw.DueDate)
		assert.Empty(t, hw.ReviewNotes)
	})

	t.Run("Empty title fails", func(t *testing.T) {
		hw, err := domain.NewHomeworkAssignment(
			clientID,
			teacherID,
			nil,
			"   ",
			"",
			nil,
		)
		assert.ErrorIs(t, err, domain.ErrEmptyHomeworkTitle)
		assert.Nil(t, hw)
	})

	t.Run("Nil client or teacher fails", func(t *testing.T) {
		_, errClient := domain.NewHomeworkAssignment(
			uuid.Nil,
			teacherID,
			nil,
			"Задание",
			"",
			nil,
		)
		assert.Error(t, errClient)

		_, errTeacher := domain.NewHomeworkAssignment(
			clientID,
			uuid.Nil,
			nil,
			"Задание",
			"",
			nil,
		)
		assert.Error(t, errTeacher)
	})

	t.Run("Status transitions", func(t *testing.T) {
		hw, err := domain.NewHomeworkAssignment(
			clientID,
			teacherID,
			nil,
			"Задание 1",
			"",
			nil,
		)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusHomeworkAssigned, hw.Status)

		// Mark completed
		err = hw.MarkCompleted("Все верно, отличная работа!")
		require.NoError(t, err)
		assert.Equal(t, domain.StatusHomeworkCompleted, hw.Status)
		assert.Equal(t, "Все верно, отличная работа!", hw.ReviewNotes)

		// Mark not done
		err = hw.MarkNotDone("Не приступал к работе")
		require.NoError(t, err)
		assert.Equal(t, domain.StatusHomeworkNotDone, hw.Status)
		assert.Equal(t, "Не приступал к работе", hw.ReviewNotes)

		// Invalid status
		err = hw.UpdateStatus(domain.HomeworkStatus("invalid"), "")
		assert.ErrorIs(t, err, domain.ErrInvalidHomeworkStatus)
	})
}
