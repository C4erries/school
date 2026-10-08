package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/domain"
)

func TestLessonJournal_Validation(t *testing.T) {
	lessonID := uuid.New()
	clientID := uuid.New()
	teacherID := uuid.New()
	validScore := 5

	t.Run("Valid creation with score", func(t *testing.T) {
		journal, err := domain.NewLessonJournal(lessonID, clientID, teacherID, "Квадратные уравнения", "Хорошо усвоил дискриминант", &validScore)
		require.NoError(t, err)
		assert.Equal(t, "Квадратные уравнения", journal.Topic)
		assert.Equal(t, "Хорошо усвоил дискриминант", journal.Notes)
		require.NotNil(t, journal.PerformanceScore)
		assert.Equal(t, 5, *journal.PerformanceScore)
	})

	t.Run("Valid creation without score", func(t *testing.T) {
		journal, err := domain.NewLessonJournal(lessonID, clientID, teacherID, "Вводный урок", "", nil)
		require.NoError(t, err)
		assert.Nil(t, journal.PerformanceScore)
	})

	t.Run("Empty topic fails", func(t *testing.T) {
		journal, err := domain.NewLessonJournal(lessonID, clientID, teacherID, "   ", "", nil)
		assert.ErrorIs(t, err, domain.ErrEmptyTopic)
		assert.Nil(t, journal)
	})

	t.Run("Score out of bounds fails", func(t *testing.T) {
		lowScore := 0
		_, errLow := domain.NewLessonJournal(lessonID, clientID, teacherID, "Тема", "", &lowScore)
		assert.ErrorIs(t, errLow, domain.ErrInvalidPerformanceScore)

		highScore := 6
		_, errHigh := domain.NewLessonJournal(lessonID, clientID, teacherID, "Тема", "", &highScore)
		assert.ErrorIs(t, errHigh, domain.ErrInvalidPerformanceScore)
	})

	t.Run("Update fields and score validation", func(t *testing.T) {
		journal, err := domain.NewLessonJournal(lessonID, clientID, teacherID, "Тема 1", "Заметка 1", &validScore)
		require.NoError(t, err)

		newScore := 4
		err = journal.Update("Тема 2", "Новая заметка", &newScore)
		require.NoError(t, err)
		assert.Equal(t, "Тема 2", journal.Topic)
		assert.Equal(t, "Новая заметка", journal.Notes)
		assert.Equal(t, 4, *journal.PerformanceScore)

		badScore := 10
		err = journal.Update("Тема 3", "", &badScore)
		assert.ErrorIs(t, err, domain.ErrInvalidPerformanceScore)

		err = journal.Update("   ", "", &newScore)
		assert.ErrorIs(t, err, domain.ErrEmptyTopic)
	})
}
