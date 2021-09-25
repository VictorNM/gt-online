package gterr

import (
	"errors"
	"fmt"
	"strings"
)

type ErrorCode string

const (
	OK                 ErrorCode = "OK"
	Cancelled          ErrorCode = "CANCELLED"
	Unknown            ErrorCode = "UNKNOWN"
	InvalidArgument    ErrorCode = "INVALID_ARGUMENT"
	DeadlineExceeded   ErrorCode = "DEADLINE_EXCEEDED"
	NotFound           ErrorCode = "NOT_FOUND"
	AlreadyExists      ErrorCode = "ALREADY_EXISTS"
	PermissionDenied   ErrorCode = "PERMISSION_DENIED"
	ResourceExhausted  ErrorCode = "RESOURCE_EXHAUSTED"
	FailedPrecondition ErrorCode = "FAILED_PRECONDITION"
	Aborted            ErrorCode = "ABORTED"
	OutOfRange         ErrorCode = "OUT_OF_RANGE"
	Unimplemented      ErrorCode = "UNIMPLEMENTED"
	Internal           ErrorCode = "INTERNAL"
	Unavailable        ErrorCode = "UNAVAILABLE"
	DataLoss           ErrorCode = "DATA_LOSS"
	Unauthenticated    ErrorCode = "UNAUTHENTICATED"
)

type multiErrors []error

func (e multiErrors) Error() string {
	b := strings.Builder{}
	for _, err := range e {
		b.WriteString("[")
		b.WriteString(err.Error())
		b.WriteString("]")
	}
	return b.String()
}

type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Detail  error     `json:"-"`
}

func (err *Error) Error() string {
	if err.Detail == nil {
		return fmt.Sprintf("code = %s, message = %s", err.Code, err.Message)
	}
	return fmt.Sprintf("code = %s, message = %s: %v", err.Code, err.Message, err.Detail)
}

func New(code ErrorCode, message string, details ...error) *Error {
	if message == "" {
		message = string(code)
	}
	e := &Error{
		Code:    code,
		Message: message,
	}
	if len(details) == 1 {
		e.Detail = details[0]
	} else if len(details) > 1 {
		e.Detail = multiErrors(details)
	}

	return e
}

func Code(err error) ErrorCode {
	return Convert(err).Code
}

func Convert(err error) *Error {
	e, _ := FromError(err)
	return e
}

func FromError(err error) (*Error, bool) {
	if err == nil {
		return New(OK, ""), false
	}

	if e := new(Error); errors.As(err, &e) {
		return e, true
	}

	return New(Unknown, ""), false
}
