package finance

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/application/finance"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/postgres"
)

// PaymentFilter алиас для finance.PaymentFilter.
type PaymentFilter = finance.PaymentFilter

// PaymentRepository реализует хранение платежей в PostgreSQL.
type PaymentRepository struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *PaymentRepository) getDBTX(ctx context.Context) postgres.DBTX {
	if tx := postgres.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

// Create сохраняет запись о платеже в БД.
func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	var notes *string
	if p.Notes != "" {
		notes = &p.Notes
	}

	query, args, err := r.sb.Insert("payments").
		Columns("id", "teacher_id", "client_id", "amount", "hours", "format", "payment_method", "paid_at", "notes", "created_at").
		Values(p.ID, p.TeacherID, p.ClientID, p.Amount, p.Hours, string(p.Format), string(p.PaymentMethod), p.PaidAt, notes, p.CreatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert payment query: %w", err)
	}

	_, err = r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec insert payment: %w", err)
	}

	return nil
}

// GetByID находит платеж по его идентификатору.
func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	query, args, err := r.sb.Select("id", "teacher_id", "client_id", "amount", "hours", "format", "payment_method", "paid_at", "notes", "created_at").
		From("payments").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select payment by id query: %w", err)
	}

	var (
		p         domain.Payment
		fmtStr    string
		methodStr string
		notes     sql.NullString
	)

	err = r.getDBTX(ctx).QueryRowContext(ctx, query, args...).Scan(
		&p.ID,
		&p.TeacherID,
		&p.ClientID,
		&p.Amount,
		&p.Hours,
		&fmtStr,
		&methodStr,
		&p.PaidAt,
		&notes,
		&p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("query payment by id: %w", err)
	}

	p.Format = domain.SubscriptionFormat(fmtStr)
	p.PaymentMethod = domain.PaymentMethod(methodStr)
	if notes.Valid {
		p.Notes = notes.String
	}

	return &p, nil
}

// List возвращает список платежей с применением фильтров и сортировкой по дате (сначала новые).
func (r *PaymentRepository) List(ctx context.Context, filter PaymentFilter) ([]*domain.Payment, error) {
	builder := r.sb.Select("id", "teacher_id", "client_id", "amount", "hours", "format", "payment_method", "paid_at", "notes", "created_at").
		From("payments").
		OrderBy("paid_at DESC", "created_at DESC")

	if filter.TeacherID != nil {
		builder = builder.Where(sq.Eq{"teacher_id": *filter.TeacherID})
	}
	if filter.ClientID != nil {
		builder = builder.Where(sq.Eq{"client_id": *filter.ClientID})
	}
	if filter.From != nil {
		builder = builder.Where(sq.GtOrEq{"paid_at": *filter.From})
	}
	if filter.To != nil {
		builder = builder.Where(sq.LtOrEq{"paid_at": *filter.To})
	}
	if filter.Limit != nil && *filter.Limit > 0 {
		builder = builder.Limit(uint64(*filter.Limit))
	}
	if filter.Offset != nil && *filter.Offset > 0 {
		builder = builder.Offset(uint64(*filter.Offset))
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list payments query: %w", err)
	}

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query payments: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		var (
			p         domain.Payment
			fmtStr    string
			methodStr string
			notes     sql.NullString
		)
		if err := rows.Scan(
			&p.ID,
			&p.TeacherID,
			&p.ClientID,
			&p.Amount,
			&p.Hours,
			&fmtStr,
			&methodStr,
			&p.PaidAt,
			&notes,
			&p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan payment: %w", err)
		}

		p.Format = domain.SubscriptionFormat(fmtStr)
		p.PaymentMethod = domain.PaymentMethod(methodStr)
		if notes.Valid {
			p.Notes = notes.String
		}
		payments = append(payments, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows payments: %w", err)
	}

	return payments, nil
}

// SumAmountByPeriod возвращает общую сумму поступивших оплат преподавателя за период.
func (r *PaymentRepository) SumAmountByPeriod(ctx context.Context, teacherID uuid.UUID, from, to time.Time) (float64, error) {
	query, args, err := r.sb.Select("COALESCE(SUM(amount), 0)").
		From("payments").
		Where(sq.Eq{"teacher_id": teacherID}).
		Where(sq.GtOrEq{"paid_at": from}).
		Where(sq.LtOrEq{"paid_at": to}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("build sum payments query: %w", err)
	}

	var sum float64
	err = r.getDBTX(ctx).QueryRowContext(ctx, query, args...).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("query sum payments: %w", err)
	}

	return sum, nil
}
