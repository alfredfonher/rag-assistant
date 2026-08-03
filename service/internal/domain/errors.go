package domain

import "errors"

var (
	ErrNotFound    = errors.New("not found")
	ErrDuplicate   = errors.New("duplicate id")
	ErrPersistence = errors.New("persistence unavailable")
)
