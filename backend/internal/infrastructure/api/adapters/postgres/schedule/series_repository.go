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

// SeriesRepository реализует операции с регулярными сериями занятий в PostgreSQL.
type SeriesRepository struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

func NewSeriesRepository(db *sql.DB) *SeriesRepository {
	return &SeriesRepository{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *SeriesRepository) getDBTX(ctx context.Context) postgres.DBTX {
	if tx := postgres.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

func (r *SeriesRepository) Create(ctx context.Context, s *domain.LessonSeries) error {
	query, args, err := r.sb.Insert("lesson_series").
		Columns(
			"id",
			"teacher_id",
			"client_id",
			"classroom_id",
			"title",
			"rrule",
			"start_time_of_day",
			"duration_minutes",
			"format",
			"location_or_url",
			"notes",
			"start_date",
			"until_date",
			"created_at",
			"updated_at",
		).
		Values(
			s.ID,
			s.TeacherID,
			s.ClientID,
			s.ClassroomID,
			s.Title,
			s.RRULE,
			s.StartTimeOfDay,
			s.DurationMinutes,
			string(s.Format),
			s.LocationOrURL,
			s.Notes,
			s.StartDate,
			s.UntilDate,
			s.CreatedAt,
			s.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert lesson_series query: %w", err)
	}

	_, err = r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec insert lesson_series: %w", err)
	}

	return nil
}

func (r *SeriesRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.LessonSeries, error) {
	query, args, err := r.sb.Select(
		"id",
		"teacher_id",
		"client_id",
		"classroom_id",
		"title",
		"rrule",
		"start_time_of_day",
		"duration_minutes",
		"format",
		"location_or_url",
		"notes",
		"start_date",
		"until_date",
		"created_at",
		"updated_at",
	).
		From("lesson_series").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select lesson_series by id query: %w", err)
	}

	var s domain.LessonSeries
	var classroomID *uuid.UUID
	var formatStr string
	var loc, notes sql.NullString
	var untilDate sql.NullTime

	err = r.getDBTX(ctx).QueryRowContext(ctx, query, args...).Scan(
		&s.ID,
		&s.TeacherID,
		&s.ClientID,
		&classroomID,
		&s.Title,
		&s.RRULE,
		&s.StartTimeOfDay,
		&s.DurationMinutes,
		&formatStr,
		&loc,
		&notes,
		&s.StartDate,
		&untilDate,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrLessonSeriesNotFound
		}
		return nil, fmt.Errorf("query lesson_series by id: %w", err)
	}

	s.ClassroomID = classroomID
	s.Format = domain.LessonFormat(formatStr)
	if loc.Valid {
		s.LocationOrURL = loc.String
	}
	if notes.Valid {
		s.Notes = notes.String
	}
	if untilDate.Valid {
		s.UntilDate = &untilDate.Time
	}

	return &s, nil
}

func (r *SeriesRepository) Update(ctx context.Context, s *domain.LessonSeries) error {
	query, args, err := r.sb.Update("lesson_series").
		Set("client_id", s.ClientID).
		Set("classroom_id", s.ClassroomID).
		Set("title", s.Title).
		Set("rrule", s.RRULE).
		Set("start_time_of_day", s.StartTimeOfDay).
		Set("duration_minutes", s.DurationMinutes).
		Set("format", string(s.Format)).
		Set("location_or_url", s.LocationOrURL).
		Set("notes", s.Notes).
		Set("start_date", s.StartDate).
		Set("until_date", s.UntilDate).
		Set("updated_at", s.UpdatedAt).
		Where(sq.Eq{"id": s.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update lesson_series query: %w", err)
	}

	res, err := r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec update lesson_series: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if rows == 0 {
		return domain.ErrLessonSeriesNotFound
	}

	return nil
}

func (r *SeriesRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := r.sb.Delete("lesson_series").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete lesson_series query: %w", err)
	}

	res, err := r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec delete lesson_series: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if rows == 0 {
		return domain.ErrLessonSeriesNotFound
	}

	return nil
}

func (r *SeriesRepository) ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.LessonSeries, error) {
	query, args, err := r.sb.Select(
		"id",
		"teacher_id",
		"client_id",
		"classroom_id",
		"title",
		"rrule",
		"start_time_of_day",
		"duration_minutes",
		"format",
		"location_or_url",
		"notes",
		"start_date",
		"until_date",
		"created_at",
		"updated_at",
	).
		From("lesson_series").
		Where(sq.Eq{"teacher_id": teacherID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list lesson_series query: %w", err)
	}

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query list lesson_series: %w", err)
	}
	defer rows.Close()

	var result []*domain.LessonSeries
	for rows.Next() {
		var s domain.LessonSeries
		var classroomID *uuid.UUID
		var formatStr string
		var loc, notes sql.NullString
		var untilDate sql.NullTime

		if err := rows.Scan(
			&s.ID,
			&s.TeacherID,
			&s.ClientID,
			&classroomID,
			&s.Title,
			&s.RRULE,
			&s.StartTimeOfDay,
			&s.DurationMinutes,
			&formatStr,
			&loc,
			&notes,
			&s.StartDate,
			&untilDate,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan lesson_series: %w", err)
		}

		s.ClassroomID = classroomID
		s.Format = domain.LessonFormat(formatStr)
		if loc.Valid {
			s.LocationOrURL = loc.String
		}
		if notes.Valid {
			s.Notes = notes.String
		}
		if untilDate.Valid {
			s.UntilDate = &untilDate.Time
		}

		result = append(result, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return result, nil
}
