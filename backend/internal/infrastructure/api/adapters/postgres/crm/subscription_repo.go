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

// SubscriptionRepository реализует хранение абонементов клиентов в PostgreSQL.
type SubscriptionRepository struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *SubscriptionRepository) getDBTX(ctx context.Context) postgres.DBTX {
	if tx := postgres.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

// Create сохраняет новый абонемент в БД.
func (r *SubscriptionRepository) Create(ctx context.Context, sub *domain.ClientSubscription) error {
	query, args, err := r.sb.Insert("client_subscriptions").
		Columns("id", "client_id", "format", "balance", "created_at").
		Values(sub.ID, sub.ClientID, string(sub.Format), sub.Balance, sub.CreatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert subscription query: %w", err)
	}

	_, err = r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec insert subscription: %w", err)
	}

	return nil
}

// GetByID находит абонемент по идентификатору.
func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ClientSubscription, error) {
	query, args, err := r.sb.Select("id", "client_id", "format", "balance", "created_at").
		From("client_subscriptions").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select subscription by id query: %w", err)
	}

	var sub domain.ClientSubscription
	var formatStr string
	err = r.getDBTX(ctx).QueryRowContext(ctx, query, args...).Scan(
		&sub.ID,
		&sub.ClientID,
		&formatStr,
		&sub.Balance,
		&sub.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("query subscription by id: %w", err)
	}

	sub.Format = domain.SubscriptionFormat(formatStr)
	return &sub, nil
}

// GetByClientID возвращает все абонементы клиента.
func (r *SubscriptionRepository) GetByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.ClientSubscription, error) {
	query, args, err := r.sb.Select("id", "client_id", "format", "balance", "created_at").
		From("client_subscriptions").
		Where(sq.Eq{"client_id": clientID}).
		OrderBy("created_at ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get subscriptions by client query: %w", err)
	}

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query subscriptions by client: %w", err)
	}
	defer rows.Close()

	var subs []*domain.ClientSubscription
	for rows.Next() {
		var sub domain.ClientSubscription
		var formatStr string
		if err := rows.Scan(
			&sub.ID,
			&sub.ClientID,
			&formatStr,
			&sub.Balance,
			&sub.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		sub.Format = domain.SubscriptionFormat(formatStr)
		subs = append(subs, &sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return subs, nil
}

// Update обновляет формат и баланс абонемента.
func (r *SubscriptionRepository) Update(ctx context.Context, sub *domain.ClientSubscription) error {
	query, args, err := r.sb.Update("client_subscriptions").
		Set("format", string(sub.Format)).
		Set("balance", sub.Balance).
		Where(sq.Eq{"id": sub.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update subscription query: %w", err)
	}

	res, err := r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec update subscription: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if rows == 0 {
		return domain.ErrSubscriptionNotFound
	}

	return nil
}

// UpdateBalance обновляет только баланс абонемента.
func (r *SubscriptionRepository) UpdateBalance(ctx context.Context, id uuid.UUID, balance float64) error {
	query, args, err := r.sb.Update("client_subscriptions").
		Set("balance", balance).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update subscription balance query: %w", err)
	}

	res, err := r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec update subscription balance: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if rows == 0 {
		return domain.ErrSubscriptionNotFound
	}

	return nil
}

// Delete удаляет абонемент по идентификатору.
func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := r.sb.Delete("client_subscriptions").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete subscription query: %w", err)
	}

	res, err := r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec delete subscription: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if rows == 0 {
		return domain.ErrSubscriptionNotFound
	}

	return nil
}
