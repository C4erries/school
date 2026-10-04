package domain

import "errors"

// Базовые доменные ошибки.
var (
	ErrNotFound         = errors.New("entity not found")
	ErrAlreadyExists    = errors.New("entity already exists")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrInternal         = errors.New("internal domain error")
)
