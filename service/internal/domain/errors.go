package domain

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrDuplicate          = errors.New("duplicate id")
	ErrPersistence        = errors.New("persistence unavailable")
	ErrAgentNotFound      = errors.New("agent not found")
	ErrCollectionNotFound = errors.New("collection not found")
	ErrAgentInUse         = errors.New("agent in use")
	ErrCollectionInUse    = errors.New("collection in use")
)
