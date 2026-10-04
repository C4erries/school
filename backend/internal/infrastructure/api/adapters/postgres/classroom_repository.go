package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
)

// ClassroomRepository реализует работу с аудиториями/кабинетами в PostgreSQL.
type ClassroomRepository struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

func NewClassroomRepository(db *sql.DB) *ClassroomRepository {
	return &ClassroomRepository{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *ClassroomRepository) getDBTX(ctx context.Context) DBTX {
	if tx := ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

// Create добавляет новый кабинет.
func (r *ClassroomRepository) Create(ctx context.Context, c *domain.Classroom) error {
	query, args, err := r.sb.Insert("classrooms").
		Columns("id", "name", "capacity", "color", "description", "created_at").
		Values(c.ID, c.Name, c.Capacity, c.Color, c.Description, c.CreatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert classroom query: %w", err)
	}

	_, err = r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec insert classroom: %w", err)
	}

	return nil
}

// GetByID возвращает кабинет по идентификатору.
func (r *ClassroomRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Classroom, error) {
	query, args, err := r.sb.Select("id", "name", "capacity", "color", "description", "created_at").
		From("classrooms").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select classroom by id query: %w", err)
	}

	var c domain.Classroom
	var desc sql.NullString
	err = r.getDBTX(ctx).QueryRowContext(ctx, query, args...).Scan(
		&c.ID,
		&c.Name,
		&c.Capacity,
		&c.Color,
		&desc,
		&c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrClassroomNotFound
		}
		return nil, fmt.Errorf("query classroom by id: %w", err)
	}

	if desc.Valid {
		c.Description = desc.String
	}

	return &c, nil
}

// List возвращает все кабинеты, отсортированные по имени.
func (r *ClassroomRepository) List(ctx context.Context) ([]*domain.Classroom, error) {
	query, args, err := r.sb.Select("id", "name", "capacity", "color", "description", "created_at").
		From("classrooms").
		OrderBy("name ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list classrooms query: %w", err)
	}

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query list classrooms: %w", err)
	}
	defer rows.Close()

	var classrooms []*domain.Classroom
	for rows.Next() {
		var c domain.Classroom
		var desc sql.NullString
		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Capacity,
			&c.Color,
			&desc,
			&c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan classroom: %w", err)
		}
		if desc.Valid {
			c.Description = desc.String
		}
		classrooms = append(classrooms, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return classrooms, nil
}

// Update обновляет информацию о кабинете.
func (r *ClassroomRepository) Update(ctx context.Context, c *domain.Classroom) error {
	query, args, err := r.sb.Update("classrooms").
		Set("name", c.Name).
		Set("capacity", c.Capacity).
		Set("color", c.Color).
		Set("description", c.Description).
		Where(sq.Eq{"id": c.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update classroom query: %w", err)
	}

	res, err := r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec update classroom: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if rows == 0 {
		return domain.ErrClassroomNotFound
	}

	return nil
}

// Delete удаляет кабинет по идентификатору.
func (r *ClassroomRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := r.sb.Delete("classrooms").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete classroom query: %w", err)
	}

	res, err := r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec delete classroom: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if rows == 0 {
		return domain.ErrClassroomNotFound
	}

	return nil
}
