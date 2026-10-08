package domain_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/domain"
)

func TestClientNote_Validation(t *testing.T) {
	clientID := uuid.New()
	teacherID := uuid.New()

	t.Run("Valid creation with trimmed content", func(t *testing.T) {
		note, err := domain.NewClientNote(clientID, teacherID, "  Позвонила мама, просила повторить дроби.  ")
		require.NoError(t, err)
		require.NotNil(t, note)
		assert.NotEqual(t, uuid.Nil, note.ID)
		assert.Equal(t, clientID, note.ClientID)
		assert.Equal(t, teacherID, note.TeacherID)
		assert.Equal(t, "Позвонила мама, просила повторить дроби.", note.Content)
		assert.False(t, note.CreatedAt.IsZero())
		assert.False(t, note.UpdatedAt.IsZero())
	})

	t.Run("Empty content fails", func(t *testing.T) {
		note, err := domain.NewClientNote(clientID, teacherID, "   ")
		assert.ErrorIs(t, err, domain.ErrEmptyClientNoteContent)
		assert.Nil(t, note)
	})

	t.Run("Nil clientID fails", func(t *testing.T) {
		note, err := domain.NewClientNote(uuid.Nil, teacherID, "Заметка")
		assert.Error(t, err)
		assert.Nil(t, note)
	})

	t.Run("Nil teacherID fails", func(t *testing.T) {
		note, err := domain.NewClientNote(clientID, uuid.Nil, "Заметка")
		assert.Error(t, err)
		assert.Nil(t, note)
	})

	t.Run("Update content success", func(t *testing.T) {
		note, err := domain.NewClientNote(clientID, teacherID, "Исходный текст")
		require.NoError(t, err)

		oldUpdated := note.UpdatedAt
		err = note.Update("  Обновленный текст заметки  ")
		require.NoError(t, err)
		assert.Equal(t, "Обновленный текст заметки", note.Content)
		assert.True(t, note.UpdatedAt.After(oldUpdated) || note.UpdatedAt.Equal(oldUpdated))
	})

	t.Run("Update content empty fails", func(t *testing.T) {
		note, err := domain.NewClientNote(clientID, teacherID, "Исходный текст")
		require.NoError(t, err)

		err = note.Update("   ")
		assert.ErrorIs(t, err, domain.ErrEmptyClientNoteContent)
		assert.Equal(t, "Исходный текст", note.Content)
	})

	t.Run("Long content succeeds", func(t *testing.T) {
		longText := strings.Repeat("Заметка ", 100)
		note, err := domain.NewClientNote(clientID, teacherID, longText)
		require.NoError(t, err)
		assert.Equal(t, strings.TrimSpace(longText), note.Content)
	})
}

