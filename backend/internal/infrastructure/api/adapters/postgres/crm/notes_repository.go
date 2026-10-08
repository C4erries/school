package crm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/postgres"
)

// NotesRepository реализует хранение свободных заметок по ученику в PostgreSQL.
type NotesRepository struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

// NewNotesRepository создает новый экземпляр NotesRepository.
func NewNotesRepository(db *sql.DB) *NotesRepository {
	return &NotesRepository{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *NotesRepository) getDBTX(ctx context.Context) postgres.DBTX {
	if tx := postgres.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

// Create сохраняет новую свободную заметку.
func (r *NotesRepository) Create(ctx context.Context, note *domain.ClientNote) error {
	query, args, err := r.sb.Insert("client_notes").
		Columns(
			"id",
			"client_id",
			"teacher_id",
			"content",
			"created_at",
			"updated_at",
		).
		Values(
			note.ID,
			note.ClientID,
			note.TeacherID,
			note.Content,
			note.CreatedAt,
			note.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert client note query: %w", err)
	}

	_, err = r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec insert client note: %w", err)
	}

	return nil
}

// GetByID находит свободную заметку по её идентификатору.
func (r *NotesRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ClientNote, error) {
	query, args, err := r.sb.Select(
		"id",
		"client_id",
		"teacher_id",
		"content",
		"created_at",
		"updated_at",
	).
		From("client_notes").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select client note query: %w", err)
	}

	var note domain.ClientNote
	err = r.getDBTX(ctx).QueryRowContext(ctx, query, args...).Scan(
		&note.ID,
		&note.ClientID,
		&note.TeacherID,
		&note.Content,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrClientNoteNotFound
		}
		return nil, fmt.Errorf("query client note by id: %w", err)
	}

	return &note, nil
}

// Delete удаляет свободную заметку по id и teacher_id.
func (r *NotesRepository) Delete(ctx context.Context, id, teacherID uuid.UUID) error {
	query, args, err := r.sb.Delete("client_notes").
		Where(sq.Eq{
			"id":         id,
			"teacher_id": teacherID,
		}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete client note query: %w", err)
	}

	res, err := r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec delete client note: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected delete client note: %w", err)
	}
	if rows == 0 {
		return domain.ErrClientNoteNotFound
	}

	return nil
}

// ListByClientID возвращает список всех свободных заметок по клиенту, отсортированных по дате создания (ASC).
func (r *NotesRepository) ListByClientID(ctx context.Context, clientID, teacherID uuid.UUID) ([]*domain.ClientNote, error) {
	query, args, err := r.sb.Select(
		"id",
		"client_id",
		"teacher_id",
		"content",
		"created_at",
		"updated_at",
	).
		From("client_notes").
		Where(sq.Eq{
			"client_id":  clientID,
			"teacher_id": teacherID,
		}).
		OrderBy("created_at ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list client notes query: %w", err)
	}

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query list client notes: %w", err)
	}
	defer rows.Close()

	notes := make([]*domain.ClientNote, 0)
	for rows.Next() {
		var note domain.ClientNote
		if err := rows.Scan(
			&note.ID,
			&note.ClientID,
			&note.TeacherID,
			&note.Content,
			&note.CreatedAt,
			&note.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan client note row: %w", err)
		}
		notes = append(notes, &note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iterate client notes: %w", err)
	}

	return notes, nil
}

