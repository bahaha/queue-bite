package domain

import "errors"

var (
	ErrTooManyOptimisticLockRetries error = errors.New("too many retries on optimistic lock")
	ErrVersionMismatch              error = errors.New("version mismatch")
)
