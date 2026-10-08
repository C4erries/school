package crm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/application/crm"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/postgres"
)

type ClientFilter = crm.ClientFilter

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

func (r *ClientRepository) getDBTX(ctx context.Context) postgres.DBTX {
	if tx := postgres.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

// Create сохраняет нового клиента в БД.
func (r *ClientRepository) Create(ctx context.Context, c *domain.Client) error {
	baseRate := c.BaseRate
	rateIndividual := c.RateIndividual
	if rateIndividual == 0 && baseRate > 0 {
		rateIndividual = baseRate
	}
	if baseRate == 0 && rateIndividual > 0 {
		baseRate = rateIndividual
	}
	c.BaseRate = baseRate
	c.RateIndividual = rateIndividual

	query, args, err := r.sb.Insert("clients").
		Columns(
			"id",
			"teacher_id",
			"name",
			"phone",
			"base_rate",
			"rate_individual",
			"rate_pair",
			"rate_group",
			"school_percent_tag",
			"is_archived",
			"created_at",
		).
		Values(
			c.ID,
			c.TeacherID,
			c.Name,
			c.Phone,
			c.BaseRate,
			c.RateIndividual,
			c.RatePair,
			c.RateGroup,
			c.SchoolPercentTag,
			c.IsArchived,
			c.CreatedAt,
		).
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

// GetByID находит клиента по идентификатору и подгружает агрегированные балансы и теги.
func (r *ClientRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	query, args, err := r.sb.Select(
		"id",
		"teacher_id",
		"name",
		"phone",
		"base_rate",
		"rate_individual",
		"rate_pair",
		"rate_group",
		"school_percent_tag",
		"is_archived",
		"created_at",
	).
		From("clients").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select client by id query: %w", err)
	}

	var c domain.Client
	var phone sql.NullString
	var ratePair, rateGroup sql.NullFloat64

	err = r.getDBTX(ctx).QueryRowContext(ctx, query, args...).Scan(
		&c.ID,
		&c.TeacherID,
		&c.Name,
		&phone,
		&c.BaseRate,
		&c.RateIndividual,
		&ratePair,
		&rateGroup,
		&c.SchoolPercentTag,
		&c.IsArchived,
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
	if ratePair.Valid {
		c.RatePair = &ratePair.Float64
	}
	if rateGroup.Valid {
		c.RateGroup = &rateGroup.Float64
	}
	if c.RateIndividual == 0 && c.BaseRate > 0 {
		c.RateIndividual = c.BaseRate
	}

	tags, err := r.getClientTags(ctx, c.ID)
	if err != nil {
		return nil, fmt.Errorf("get client tags: %w", err)
	}
	c.Tags = tags

	balances, err := r.getClientBalances(ctx, c.ID)
	if err != nil {
		return nil, fmt.Errorf("get client balances: %w", err)
	}
	c.Balances = balances

	lastLesson, err := r.getClientLastLesson(ctx, c.ID)
	if err != nil {
		return nil, fmt.Errorf("get client last lesson: %w", err)
	}
	c.LastLessonAt = lastLesson

	return &c, nil
}

// Update обновляет поля клиента.
func (r *ClientRepository) Update(ctx context.Context, c *domain.Client) error {
	baseRate := c.BaseRate
	rateIndividual := c.RateIndividual
	if rateIndividual == 0 && baseRate > 0 {
		rateIndividual = baseRate
	}
	if baseRate == 0 && rateIndividual > 0 {
		baseRate = rateIndividual
	}
	c.BaseRate = baseRate
	c.RateIndividual = rateIndividual

	query, args, err := r.sb.Update("clients").
		Set("name", c.Name).
		Set("phone", c.Phone).
		Set("base_rate", c.BaseRate).
		Set("rate_individual", c.RateIndividual).
		Set("rate_pair", c.RatePair).
		Set("rate_group", c.RateGroup).
		Set("school_percent_tag", c.SchoolPercentTag).
		Set("is_archived", c.IsArchived).
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

