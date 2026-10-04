package application

import "context"

// Transactor выполняет функцию fn внутри единой транзакции базы данных.
// Транзакция инжектируется в context.Context и извлекается репозиториями в слое infrastructure.
type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}
