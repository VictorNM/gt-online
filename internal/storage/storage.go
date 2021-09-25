package storage

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrAlreadyExist    = errors.New("already exist")
)
