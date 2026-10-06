package finance

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

// PartnerPayoutRepository реализует хранение выплат партнерам в PostgreSQL.
type PartnerPayoutRepository struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

func NewPartnerPayoutRepository(db *sql.DB) *PartnerPayoutRepository {
	return &PartnerPayoutRepository{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *PartnerPayoutRepository) getDBTX(ctx context.Context) postgres.DBTX {
	if tx := postgres.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

// Create сохраняет запись о выплате комиссии партнерской школе.
func (r *PartnerPayoutRepository) Create(ctx context.Context, p *domain.PartnerPayout) error {
	var notes *string
	if p.Notes != "" {
		notes = &p.Notes
	}

	query, args, err := r.sb.Insert("partner_payouts").
		Columns("id", "teacher_id", "tag_id", "period_month", "gross_amount", "commission_amount", "paid_at", "notes", "created_at").
		Values(p.ID, p.TeacherID, p.TagID, p.PeriodMonth, p.GrossAmount, p.CommissionAmount, p.PaidAt, notes, p.CreatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert partner payout query: %w", err)
	}

	_, err = r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec insert partner payout: %w", err)
	}

	return nil
}

// GetByID находит выплату по идентификатору.
func (r *PartnerPayoutRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PartnerPayout, error) {
	query, args, err := r.sb.Select("id", "teacher_id", "tag_id", "period_month", "gross_amount", "commission_amount", "paid_at", "notes", "created_at").
		From("partner_payouts").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select partner payout by id query: %w", err)
	}

	var (
		p     domain.PartnerPayout
		notes sql.NullString
	)

	err = r.getDBTX(ctx).QueryRowContext(ctx, query, args...).Scan(
		&p.ID,
		&p.TeacherID,
		&p.TagID,
		&p.PeriodMonth,
		&p.GrossAmount,
		&p.CommissionAmount,
		&p.PaidAt,
		&notes,
		&p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPartnerPayoutNotFound
		}
		return nil, fmt.Errorf("query partner payout by id: %w", err)
	}

	if notes.Valid {
		p.Notes = notes.String
	}

	return &p, nil
}

// GetByTagAndPeriod находит выплату партнерской школе по тегу и месяцу.
func (r *PartnerPayoutRepository) GetByTagAndPeriod(ctx context.Context, teacherID, tagID uuid.UUID, periodMonth string) (*domain.PartnerPayout, error) {
	query, args, err := r.sb.Select("id", "teacher_id", "tag_id", "period_month", "gross_amount", "commission_amount", "paid_at", "notes", "created_at").
		From("partner_payouts").
		Where(sq.Eq{
			"teacher_id":   teacherID,
			"tag_id":       tagID,
			"period_month": periodMonth,
		}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select partner payout by tag and period query: %w", err)
	}

	var (
		p     domain.PartnerPayout
		notes sql.NullString
	)

	err = r.getDBTX(ctx).QueryRowContext(ctx, query, args...).Scan(
		&p.ID,
		&p.TeacherID,
		&p.TagID,
		&p.PeriodMonth,
		&p.GrossAmount,
		&p.CommissionAmount,
		&p.PaidAt,
		&notes,
		&p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPartnerPayoutNotFound
		}
		return nil, fmt.Errorf("query partner payout by tag and period: %w", err)
	}

	if notes.Valid {
		p.Notes = notes.String
	}

	return &p, nil
}

// ListByPeriod возвращает все выплаты партнерам за указанный расчетный месяц.
func (r *PartnerPayoutRepository) ListByPeriod(ctx context.Context, teacherID uuid.UUID, periodMonth string) ([]*domain.PartnerPayout, error) {
	query, args, err := r.sb.Select("id", "teacher_id", "tag_id", "period_month", "gross_amount", "commission_amount", "paid_at", "notes", "created_at").
		From("partner_payouts").
		Where(sq.Eq{
			"teacher_id":   teacherID,
			"period_month": periodMonth,
		}).
		OrderBy("paid_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list partner payouts by period query: %w", err)
	}

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query partner payouts by period: %w", err)
	}
	defer rows.Close()

	var payouts []*domain.PartnerPayout
	for rows.Next() {
		var (
			p     domain.PartnerPayout
			notes sql.NullString
		)
		if err := rows.Scan(
			&p.ID,
			&p.TeacherID,
			&p.TagID,
			&p.PeriodMonth,
			&p.GrossAmount,
			&p.CommissionAmount,
			&p.PaidAt,
			&notes,
			&p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan partner payout: %w", err)
		}
		if notes.Valid {
			p.Notes = notes.String
		}
		payouts = append(payouts, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows partner payouts: %w", err)
	}

	return payouts, nil
}

// ListByTeacherID возвращает все зафиксированные выплаты преподавателя.
func (r *PartnerPayoutRepository) ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.PartnerPayout, error) {
	query, args, err := r.sb.Select("id", "teacher_id", "tag_id", "period_month", "gross_amount", "commission_amount", "paid_at", "notes", "created_at").
		From("partner_payouts").
		Where(sq.Eq{"teacher_id": teacherID}).
		OrderBy("paid_at DESC", "created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list partner payouts by teacher query: %w", err)
	}

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query partner payouts by teacher: %w", err)
	}
	defer rows.Close()

	var payouts []*domain.PartnerPayout
	for rows.Next() {
		var (
			p     domain.PartnerPayout
			notes sql.NullString
		)
		if err := rows.Scan(
			&p.ID,
			&p.TeacherID,
			&p.TagID,
			&p.PeriodMonth,
			&p.GrossAmount,
			&p.CommissionAmount,
			&p.PaidAt,
			&notes,
			&p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan partner payout: %w", err)
		}
		if notes.Valid {
			p.Notes = notes.String
		}
		payouts = append(payouts, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows partner payouts: %w", err)
	}

	return payouts, nil
}
