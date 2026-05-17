package apperr

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("already exists")
	ErrExpired  = errors.New("expired")
)
