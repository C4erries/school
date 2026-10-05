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

	// Подгружаем теги
	tags, err := r.getClientTags(ctx, c.ID)
	if err != nil {
		return nil, fmt.Errorf("get client tags: %w", err)
	}
	c.Tags = tags

	// Подгружаем балансы
	balances, err := r.getClientBalances(ctx, c.ID)
	if err != nil {
		return nil, fmt.Errorf("get client balances: %w", err)
	}
	c.Balances = balances

	return &c, nil
}

// ListByTeacherID возвращает всех клиентов указанного преподавателя с агрегированными балансами и тегами.
func (r *ClientRepository) ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.Client, error) {
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
		"created_at",
	).
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
		var ratePair, rateGroup sql.NullFloat64

		if err := rows.Scan(
			&c.ID,
			&c.TeacherID,
			&c.Name,
			&phone,
			&c.BaseRate,
			&c.RateIndividual,
			&ratePair,
			&rateGroup,
			&c.SchoolPercentTag,
			&c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan client: %w", err)
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
		c.Tags = make([]domain.Tag, 0)

		clients = append(clients, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	if len(clients) == 0 {
		return clients, nil
	}

	// Пакетная подгрузка тегов для всех клиентов преподавателя
	tagMap, err := r.getBatchClientTags(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("get batch client tags: %w", err)
	}

	// Пакетная подгрузка балансов для всех клиентов преподавателя
	balanceMap, err := r.getBatchClientBalances(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("get batch client balances: %w", err)
	}

	for _, c := range clients {
		if tags, exists := tagMap[c.ID]; exists {
			c.Tags = tags
		}
		if bal, exists := balanceMap[c.ID]; exists {
			c.Balances = bal
		}
	}

	return clients, nil
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

func (r *ClientRepository) getClientTags(ctx context.Context, clientID uuid.UUID) ([]domain.Tag, error) {
	query := `SELECT t.id, t.teacher_id, t.name, t.school_percent, t.color, t.created_at
		FROM tags t
		JOIN client_tags ct ON ct.tag_id = t.id
		WHERE ct.client_id = $1
		ORDER BY t.name ASC`

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]domain.Tag, 0)
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(
			&t.ID,
			&t.TeacherID,
			&t.Name,
			&t.SchoolPercent,
			&t.Color,
			&t.CreatedAt,
		); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (r *ClientRepository) getClientBalances(ctx context.Context, clientID uuid.UUID) (domain.ClientBalances, error) {
	query := `SELECT format, COALESCE(SUM(balance), 0)
		FROM client_subscriptions
		WHERE client_id = $1
		GROUP BY format`

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, clientID)
	if err != nil {
		return domain.ClientBalances{}, err
	}
	defer rows.Close()

	var b domain.ClientBalances
	for rows.Next() {
		var format string
		var sum float64
		if err := rows.Scan(&format, &sum); err != nil {
			return b, err
		}
		switch domain.SubscriptionFormat(format) {
		case domain.SubscriptionFormatIndividual:
			b.IndividualHours = sum
		case domain.SubscriptionFormatPair:
			b.PairHours = sum
		case domain.SubscriptionFormatGroup:
			b.GroupHours = sum
		}
	}
	b.TotalHours = b.IndividualHours + b.PairHours + b.GroupHours
	return b, rows.Err()
}

func (r *ClientRepository) getBatchClientTags(ctx context.Context, teacherID uuid.UUID) (map[uuid.UUID][]domain.Tag, error) {
	query := `SELECT ct.client_id, t.id, t.teacher_id, t.name, t.school_percent, t.color, t.created_at
		FROM tags t
		JOIN client_tags ct ON ct.tag_id = t.id
		JOIN clients c ON c.id = ct.client_id
		WHERE c.teacher_id = $1
		ORDER BY t.name ASC`

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[uuid.UUID][]domain.Tag)
	for rows.Next() {
		var clientID uuid.UUID
		var t domain.Tag
		if err := rows.Scan(
			&clientID,
			&t.ID,
			&t.TeacherID,
			&t.Name,
			&t.SchoolPercent,
			&t.Color,
			&t.CreatedAt,
		); err != nil {
			return nil, err
		}
		res[clientID] = append(res[clientID], t)
	}
	return res, rows.Err()
}

func (r *ClientRepository) getBatchClientBalances(ctx context.Context, teacherID uuid.UUID) (map[uuid.UUID]domain.ClientBalances, error) {
	query := `SELECT cs.client_id, cs.format, COALESCE(SUM(cs.balance), 0)
		FROM client_subscriptions cs
		JOIN clients c ON c.id = cs.client_id
		WHERE c.teacher_id = $1
		GROUP BY cs.client_id, cs.format`

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[uuid.UUID]domain.ClientBalances)
	for rows.Next() {
		var clientID uuid.UUID
		var format string
		var sum float64
		if err := rows.Scan(&clientID, &format, &sum); err != nil {
			return nil, err
		}
		b := res[clientID]
		switch domain.SubscriptionFormat(format) {
		case domain.SubscriptionFormatIndividual:
			b.IndividualHours = sum
		case domain.SubscriptionFormatPair:
			b.PairHours = sum
		case domain.SubscriptionFormatGroup:
			b.GroupHours = sum
		}
		b.TotalHours = b.IndividualHours + b.PairHours + b.GroupHours
		res[clientID] = b
	}
	return res, rows.Err()
}
