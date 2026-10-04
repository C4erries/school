package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/C4erries/school/backend/internal/domain"
)

// UserRepository реализует хранение пользователей в PostgreSQL.
type UserRepository struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *UserRepository) getDBTX(ctx context.Context) DBTX {
	if tx := ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}
	return false
}

// Create сохраняет нового пользователя в БД.
func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	query, args, err := r.sb.Insert("users").
		Columns("id", "email", "password_hash", "full_name", "phone", "role", "created_at", "updated_at").
		Values(u.ID, u.Email, u.PasswordHash, u.FullName, u.Phone, string(u.Role), u.CreatedAt, u.UpdatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert user query: %w", err)
	}

	_, err = r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("exec insert user: %w", err)
	}

	return nil
}

// GetByID находит пользователя по UUID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query, args, err := r.sb.Select("id", "email", "password_hash", "full_name", "phone", "role", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select user by id query: %w", err)
	}

	var u domain.User
	var role string
	err = r.getDBTX(ctx).QueryRowContext(ctx, query, args...).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&u.Phone,
		&role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("query user by id: %w", err)
	}

	u.Role = domain.Role(role)
	return &u, nil
}

// GetByEmail находит пользователя по email (без учета регистра).
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query, args, err := r.sb.Select("id", "email", "password_hash", "full_name", "phone", "role", "created_at", "updated_at").
		From("users").
		Where("LOWER(email) = LOWER(?)", email).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select user by email query: %w", err)
	}

	var u domain.User
	var role string
	err = r.getDBTX(ctx).QueryRowContext(ctx, query, args...).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&u.Phone,
		&role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("query user by email: %w", err)
	}

	u.Role = domain.Role(role)
	return &u, nil
}

// Update обновляет поля существующего пользователя.
func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	query, args, err := r.sb.Update("users").
		Set("email", u.Email).
		Set("password_hash", u.PasswordHash).
		Set("full_name", u.FullName).
		Set("phone", u.Phone).
		Set("role", string(u.Role)).
		Set("updated_at", u.UpdatedAt).
		Where(sq.Eq{"id": u.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update user query: %w", err)
	}

	res, err := r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("exec update user: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if rows == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
