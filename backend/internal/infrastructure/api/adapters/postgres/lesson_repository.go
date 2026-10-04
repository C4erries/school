package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
)

// LessonFilter параметры фильтрации списка уроков.
type LessonFilter = schedule.LessonFilter

// LessonRepository реализует операции с уроками в PostgreSQL.
type LessonRepository struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

func NewLessonRepository(db *sql.DB) *LessonRepository {
	return &LessonRepository{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *LessonRepository) getDBTX(ctx context.Context) DBTX {
	if tx := ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

// HasClassroomCollision проверяет, занят ли данный кабинет ДРУГИМ преподавателем в указанный интервал.
// Нахлёст у одного и того же преподавателя разрешен согласно бизнес-требованиям.
func (r *LessonRepository) HasClassroomCollision(
	ctx context.Context,
	classroomID uuid.UUID,
	teacherID uuid.UUID,
	startTime time.Time,
	endTime time.Time,
	excludeLessonID *uuid.UUID,
) (bool, error) {
	builder := r.sb.Select("1").
		From("lessons").
		Where(sq.Eq{"classroom_id": classroomID}).
		Where(sq.NotEq{"teacher_id": teacherID}).
		Where(sq.Expr("status NOT IN (?, ?, ?)",
			string(domain.StatusCancelledByTeacher),
			string(domain.StatusCancelledByStudent),
			string(domain.StatusDeclined),
		)).
		Where("start_time < ?", endTime).
		Where("end_time > ?", startTime)

	if excludeLessonID != nil && *excludeLessonID != uuid.Nil {
		builder = builder.Where(sq.NotEq{"id": *excludeLessonID})
	}

	query, args, err := builder.Limit(1).ToSql()
	if err != nil {
		return false, fmt.Errorf("build has classroom collision query: %w", err)
	}

	var dummy int
	err = r.getDBTX(ctx).QueryRowContext(ctx, query, args...).Scan(&dummy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query has classroom collision: %w", err)
	}

	return true, nil
}

// Create сохраняет новый урок в БД.
func (r *LessonRepository) Create(ctx context.Context, l *domain.Lesson) error {
	query, args, err := r.sb.Insert("lessons").
		Columns(
			"id",
			"teacher_id",
			"student_id",
			"classroom_id",
			"start_time",
			"end_time",
			"format",
			"location_or_url",
			"status",
			"notes",
			"cancel_reason",
			"created_at",
			"updated_at",
		).
		Values(
			l.ID,
			l.TeacherID,
			l.StudentID,
			l.ClassroomID,
			l.StartTime,
			l.EndTime,
			string(l.Format),
			l.LocationOrURL,
			string(l.Status),
			l.Notes,
			l.CancelReason,
			l.CreatedAt,
			l.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert lesson query: %w", err)
	}

	_, err = r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec insert lesson: %w", err)
	}

	return nil
}

// GetByID находит урок по идентификатору.
func (r *LessonRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Lesson, error) {
	query, args, err := r.sb.Select(
		"id",
		"teacher_id",
		"student_id",
		"classroom_id",
		"start_time",
		"end_time",
		"format",
		"location_or_url",
		"status",
		"notes",
		"cancel_reason",
		"created_at",
		"updated_at",
	).
		From("lessons").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select lesson by id query: %w", err)
	}

	var l domain.Lesson
	var classroomID *uuid.UUID
	var formatStr, statusStr string
	var loc, notes, cancelReason sql.NullString

	err = r.getDBTX(ctx).QueryRowContext(ctx, query, args...).Scan(
		&l.ID,
		&l.TeacherID,
		&l.StudentID,
		&classroomID,
		&l.StartTime,
		&l.EndTime,
		&formatStr,
		&loc,
		&statusStr,
		&notes,
		&cancelReason,
		&l.CreatedAt,
		&l.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrLessonNotFound
		}
		return nil, fmt.Errorf("query lesson by id: %w", err)
	}

	l.ClassroomID = classroomID
	l.Format = domain.LessonFormat(formatStr)
	l.Status = domain.LessonStatus(statusStr)
	if loc.Valid {
		l.LocationOrURL = loc.String
	}
	if notes.Valid {
		l.Notes = notes.String
	}
	if cancelReason.Valid {
		l.CancelReason = cancelReason.String
	}

	return &l, nil
}

// Update обновляет поля урока.
func (r *LessonRepository) Update(ctx context.Context, l *domain.Lesson) error {
	query, args, err := r.sb.Update("lessons").
		Set("teacher_id", l.TeacherID).
		Set("student_id", l.StudentID).
		Set("classroom_id", l.ClassroomID).
		Set("start_time", l.StartTime).
		Set("end_time", l.EndTime).
		Set("format", string(l.Format)).
		Set("location_or_url", l.LocationOrURL).
		Set("status", string(l.Status)).
		Set("notes", l.Notes).
		Set("cancel_reason", l.CancelReason).
		Set("updated_at", l.UpdatedAt).
		Where(sq.Eq{"id": l.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update lesson query: %w", err)
	}

	res, err := r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec update lesson: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if rows == 0 {
		return domain.ErrLessonNotFound
	}

	return nil
}

// List возвращает список уроков согласно фильтрам, упорядоченный по времени начала.
func (r *LessonRepository) List(ctx context.Context, filter LessonFilter) ([]*domain.Lesson, error) {
	builder := r.sb.Select(
		"id",
		"teacher_id",
		"student_id",
		"classroom_id",
		"start_time",
		"end_time",
		"format",
		"location_or_url",
		"status",
		"notes",
		"cancel_reason",
		"created_at",
		"updated_at",
	).From("lessons")

	if filter.TeacherID != nil {
		builder = builder.Where(sq.Eq{"teacher_id": *filter.TeacherID})
	}
	if filter.StudentID != nil {
		builder = builder.Where(sq.Eq{"student_id": *filter.StudentID})
	}
	if filter.ClassroomID != nil {
		builder = builder.Where(sq.Eq{"classroom_id": *filter.ClassroomID})
	}
	if filter.Status != nil {
		builder = builder.Where(sq.Eq{"status": string(*filter.Status)})
	}
	if filter.From != nil {
		builder = builder.Where("end_time >= ?", *filter.From)
	}
	if filter.To != nil {
		builder = builder.Where("start_time <= ?", *filter.To)
	}

	query, args, err := builder.OrderBy("start_time ASC").ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list lessons query: %w", err)
	}

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query list lessons: %w", err)
	}
	defer rows.Close()

	var lessons []*domain.Lesson
	for rows.Next() {
		var l domain.Lesson
		var classroomID *uuid.UUID
		var formatStr, statusStr string
		var loc, notes, cancelReason sql.NullString

		if err := rows.Scan(
			&l.ID,
			&l.TeacherID,
			&l.StudentID,
			&classroomID,
			&l.StartTime,
			&l.EndTime,
			&formatStr,
			&loc,
			&statusStr,
			&notes,
			&cancelReason,
			&l.CreatedAt,
			&l.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan lesson: %w", err)
		}

		l.ClassroomID = classroomID
		l.Format = domain.LessonFormat(formatStr)
		l.Status = domain.LessonStatus(statusStr)
		if loc.Valid {
			l.LocationOrURL = loc.String
		}
		if notes.Valid {
			l.Notes = notes.String
		}
		if cancelReason.Valid {
			l.CancelReason = cancelReason.String
		}

		lessons = append(lessons, &l)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return lessons, nil
}
