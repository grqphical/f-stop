package database

import "errors"

var (
	ErrUniqueConstraint = errors.New("unique constraint failed")
)
