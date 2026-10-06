package crm

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/postgres"
)

// BalanceAdjustmentRepository реализует хранение аудита ручных корректировок баланса в PostgreSQL.
type BalanceAdjustmentRepository struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

func NewBalanceAdjustmentRepository(db *sql.DB) *BalanceAdjustmentRepository {
	return &BalanceAdjustmentRepository{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *BalanceAdjustmentRepository) getDBTX(ctx context.Context) postgres.DBTX {
	if tx := postgres.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

// Create сохраняет запись о ручной корректировке баланса.
func (r *BalanceAdjustmentRepository) Create(ctx context.Context, adj *domain.ClientBalanceAdjustment) error {
	query, args, err := r.sb.Insert("client_balance_adjustments").
		Columns("id", "client_id", "teacher_id", "format", "delta_hours", "reason", "created_at").
		Values(adj.ID, adj.ClientID, adj.TeacherID, string(adj.Format), adj.DeltaHours, adj.Reason, adj.CreatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert balance adjustment query: %w", err)
	}

	_, err = r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec insert balance adjustment: %w", err)
	}

	return nil
}

// ListByClientID возвращает все записи аудита корректировок баланса для указанного клиента.
func (r *BalanceAdjustmentRepository) ListByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.ClientBalanceAdjustment, error) {
	query, args, err := r.sb.Select("id", "client_id", "teacher_id", "format", "delta_hours", "reason", "created_at").
		From("client_balance_adjustments").
		Where(sq.Eq{"client_id": clientID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list balance adjustments query: %w", err)
	}

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query balance adjustments: %w", err)
	}
	defer rows.Close()

	var adjustments []*domain.ClientBalanceAdjustment
	for rows.Next() {
		var (
			adj    domain.ClientBalanceAdjustment
			format string
		)
		if err := rows.Scan(
			&adj.ID,
			&adj.ClientID,
			&adj.TeacherID,
			&format,
			&adj.DeltaHours,
			&adj.Reason,
			&adj.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan balance adjustment: %w", err)
		}
		adj.Format = domain.SubscriptionFormat(format)
		adjustments = append(adjustments, &adj)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows balance adjustments: %w", err)
	}

	return adjustments, nil
}
