package transaction

import "errors"

var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrTransactionNotFound     = errors.New("transaction not found")
)
