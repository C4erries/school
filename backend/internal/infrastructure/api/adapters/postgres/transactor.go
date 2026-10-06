package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type txKey struct{}

// InjectTx помещает транзакцию в контекст.
func InjectTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// ExtractTx извлекает транзакцию из контекста, если она есть.
func ExtractTx(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return nil
}

// DBTX обобщает методы выполнения SQL для *sql.DB и *sql.Tx.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Transactor реализует application.Transactor для PostgreSQL.
type Transactor struct {
	db *sql.DB
}

func NewTransactor(db *sql.DB) *Transactor {
	return &Transactor{db: db}
}

// WithinTransaction выполняет fn внутри транзакции. Если транзакция уже существует в контексте, fn выполняется в ней.
func (t *Transactor) WithinTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	if ExtractTx(ctx) != nil {
		return fn(ctx)
	}

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	txCtx := InjectTx(ctx, tx)

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction err: %w, rollback err: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// IsUniqueViolation проверяет, является ли ошибка нарушением уникальности PostgreSQL (код 23505).
func IsUniqueViolation(err error) bool {
	type pgError interface {
		SQLState() string
	}
	var pgErr pgError
	if errors.As(err, &pgErr) && pgErr.SQLState() == "23505" {
		return true
	}
	return false
}
