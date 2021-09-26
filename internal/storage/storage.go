package storage

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrAlreadyExist    = errors.New("already exist")
)

func IsErrNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func IsErrInvalidArgument(err error) bool {
	return errors.Is(err, ErrInvalidArgument)
}

func IsErrAlreadyExist(err error) bool {
	return errors.Is(err, ErrAlreadyExist)
}
