package schedule

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

// JournalRepository реализует сохранение отчетов по урокам в PostgreSQL.
type JournalRepository struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

// NewJournalRepository создает новый экземпляр JournalRepository.
func NewJournalRepository(db *sql.DB) *JournalRepository {
	return &JournalRepository{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *JournalRepository) getDBTX(ctx context.Context) postgres.DBTX {
	if tx := postgres.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

// GetByLessonID возвращает отчет по идентификатору урока.
func (r *JournalRepository) GetByLessonID(ctx context.Context, lessonID uuid.UUID) (*domain.LessonJournal, error) {
	query, args, err := r.sb.Select(
		"id",
		"lesson_id",
		"client_id",
		"teacher_id",
		"topic",
		"notes",
		"performance_score",
		"created_at",
		"updated_at",
	).
		From("lesson_journals").
		Where(sq.Eq{"lesson_id": lessonID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select journal query: %w", err)
	}

	var j domain.LessonJournal
	var notes sql.NullString
	var score sql.NullInt32

	err = r.getDBTX(ctx).QueryRowContext(ctx, query, args...).Scan(
		&j.ID,
		&j.LessonID,
		&j.ClientID,
		&j.TeacherID,
		&j.Topic,
		&notes,
		&score,
		&j.CreatedAt,
		&j.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrJournalNotFound
		}
		return nil, fmt.Errorf("scan lesson journal: %w", err)
	}

	if notes.Valid {
		j.Notes = notes.String
	}
	if score.Valid {
		s := int(score.Int32)
		j.PerformanceScore = &s
	}

	return &j, nil
}

// Upsert создает или обновляет отчет по уроку с возвратом сохраненных полей.
func (r *JournalRepository) Upsert(ctx context.Context, journal *domain.LessonJournal) error {
	const rawQuery = `
		INSERT INTO lesson_journals (
			id, lesson_id, client_id, teacher_id, topic, notes, performance_score, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (lesson_id) DO UPDATE SET
			topic = EXCLUDED.topic,
			notes = EXCLUDED.notes,
			performance_score = EXCLUDED.performance_score,
			updated_at = EXCLUDED.updated_at
		RETURNING id, created_at, updated_at;
	`

	var scoreVal *int
	if journal.PerformanceScore != nil {
		scoreVal = journal.PerformanceScore
	}

	err := r.getDBTX(ctx).QueryRowContext(
		ctx,
		rawQuery,
		journal.ID,
		journal.LessonID,
		journal.ClientID,
		journal.TeacherID,
		journal.Topic,
		journal.Notes,
		scoreVal,
		journal.CreatedAt,
		journal.UpdatedAt,
	).Scan(&journal.ID, &journal.CreatedAt, &journal.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert lesson journal: %w", err)
	}

	return nil
}

// ListByClientID возвращает все отчеты по урокам ученика в хронологическом порядке.
func (r *JournalRepository) ListByClientID(ctx context.Context, clientID, teacherID uuid.UUID) ([]*domain.LessonJournal, error) {
	query, args, err := r.sb.Select(
		"id",
		"lesson_id",
		"client_id",
		"teacher_id",
		"topic",
		"notes",
		"performance_score",
		"created_at",
		"updated_at",
	).
		From("lesson_journals").
		Where(sq.Eq{
			"client_id":  clientID,
			"teacher_id": teacherID,
		}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list client journals query: %w", err)
	}

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query list client journals: %w", err)
	}
	defer rows.Close()

	journals := make([]*domain.LessonJournal, 0)
	for rows.Next() {
		var j domain.LessonJournal
		var notes sql.NullString
		var score sql.NullInt32

		if err := rows.Scan(
			&j.ID,
			&j.LessonID,
			&j.ClientID,
			&j.TeacherID,
			&j.Topic,
			&notes,
			&score,
			&j.CreatedAt,
			&j.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan client journal row: %w", err)
		}

		if notes.Valid {
			j.Notes = notes.String
		}
		if score.Valid {
			s := int(score.Int32)
			j.PerformanceScore = &s
		}
		journals = append(journals, &j)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iterate journals: %w", err)
	}

	return journals, nil
}
