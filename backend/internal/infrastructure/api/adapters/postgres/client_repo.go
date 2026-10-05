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

// ClientRepository реализует хранение клиентов репетитора в PostgreSQL.
type ClientRepository struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

func NewClientRepository(db *sql.DB) *ClientRepository {
	return &ClientRepository{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *ClientRepository) getDBTX(ctx context.Context) DBTX {
	if tx := ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

// Create сохраняет нового клиента в БД.
func (r *ClientRepository) Create(ctx context.Context, c *domain.Client) error {
	query, args, err := r.sb.Insert("clients").
		Columns("id", "teacher_id", "name", "phone", "base_rate", "school_percent_tag", "created_at").
		Values(c.ID, c.TeacherID, c.Name, c.Phone, c.BaseRate, c.SchoolPercentTag, c.CreatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert client query: %w", err)
	}

	_, err = r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec insert client: %w", err)
	}

	return nil
}

// GetByID находит клиента по идентификатору.
func (r *ClientRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	query, args, err := r.sb.Select("id", "teacher_id", "name", "phone", "base_rate", "school_percent_tag", "created_at").
		From("clients").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select client by id query: %w", err)
	}

	var c domain.Client
	var phone sql.NullString
	err = r.getDBTX(ctx).QueryRowContext(ctx, query, args...).Scan(
		&c.ID,
		&c.TeacherID,
		&c.Name,
		&phone,
		&c.BaseRate,
		&c.SchoolPercentTag,
		&c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrClientNotFound
		}
		return nil, fmt.Errorf("query client by id: %w", err)
	}

	if phone.Valid {
		c.Phone = &phone.String
	}

	return &c, nil
}

// ListByTeacherID возвращает всех клиентов указанного преподавателя.
func (r *ClientRepository) ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.Client, error) {
	query, args, err := r.sb.Select("id", "teacher_id", "name", "phone", "base_rate", "school_percent_tag", "created_at").
		From("clients").
		Where(sq.Eq{"teacher_id": teacherID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list clients query: %w", err)
	}

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query list clients: %w", err)
	}
	defer rows.Close()

	var clients []*domain.Client
	for rows.Next() {
		var c domain.Client
		var phone sql.NullString
		if err := rows.Scan(
			&c.ID,
			&c.TeacherID,
			&c.Name,
			&phone,
			&c.BaseRate,
			&c.SchoolPercentTag,
			&c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan client: %w", err)
		}
		if phone.Valid {
			c.Phone = &phone.String
		}
		clients = append(clients, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return clients, nil
}

// Update обновляет поля клиента.
func (r *ClientRepository) Update(ctx context.Context, c *domain.Client) error {
	query, args, err := r.sb.Update("clients").
		Set("name", c.Name).
		Set("phone", c.Phone).
		Set("base_rate", c.BaseRate).
		Set("school_percent_tag", c.SchoolPercentTag).
		Where(sq.Eq{"id": c.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update client query: %w", err)
	}

	res, err := r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec update client: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if rows == 0 {
		return domain.ErrClientNotFound
	}

	return nil
}

// Delete удаляет клиента по идентификатору.
func (r *ClientRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := r.sb.Delete("clients").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete client query: %w", err)
	}

	res, err := r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec delete client: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if rows == 0 {
		return domain.ErrClientNotFound
	}

	return nil
}
