package apperrors

import (
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/google/uuid"
)

type Code int

const (
	InternalCode Code = iota + 1
	BadRequestCode
	NotFoundCode
	AlreadyExistsCode
	UnauthorizedCode
	ForbiddenCode
)

var _ error = (*Error)(nil)

type Error struct {
	Code       Code
	StackTrace string
	IncidentID string

	innerErr error
	hideErr  bool
	message  string
}

func (e *Error) Error() string {
	return e.error(false)
}

func (e *Error) SafeError() string {
	return e.error(true)
}

func (e *Error) Unwrap() error {
	return e.innerErr
}

func (e *Error) error(safe bool) string {
	switch {
	case e.innerErr == nil:
		return e.message
	case safe && e.hideErr:
		return e.message
	case e.message == "":
		return e.innerErr.Error()
	default:
		return fmt.Sprintf("%s: %s", e.message, e.innerErr.Error())
	}
}

func Internal(err error) *Error {
	appErr := InternalWithoutStackTrace(err)
	appErr.StackTrace = string(debug.Stack())
	return appErr
}

func InternalWithoutStackTrace(err error) *Error {
	appErr := newHiddenError(err, InternalCode, "internal error")
	appErr.IncidentID = uuid.New().String()
	return appErr
}

func BadRequest(err error) *Error {
	return newWrappedError(err, BadRequestCode)
}

func BadRequestHidden(err error, message string) *Error {
	return newHiddenError(err, BadRequestCode, message)
}

func NotFound(subject, key string, value any) *Error {
	return newError(NotFoundCode, fmt.Sprintf("%s %s:%v not found", subject, key, value))
}

func AlreadyExists(subject, key string, value any) *Error {
	return newError(AlreadyExistsCode, fmt.Sprintf("%s %s:%v already exists", subject, key, value))
}

func Unauthorized(message string) *Error {
	return newError(UnauthorizedCode, message)
}

func UnauthorizedHidden(err error, message string) *Error {
	return newHiddenError(err, UnauthorizedCode, message)
}

func Forbidden(message string) *Error {
	return newError(ForbiddenCode, message)
}

func Is(err error, code Code) bool {
	var appErr *Error
	return errors.As(err, &appErr) && appErr.Code == code
}

func newError(code Code, message string) *Error {
	return &Error{
		Code:    code,
		message: message,
	}
}

func newWrappedError(err error, code Code) *Error {
	return &Error{
		Code:     code,
		innerErr: err,
	}
}

func newHiddenError(err error, code Code, message string) *Error {
	return &Error{
		Code:     code,
		message:  message,
		innerErr: err,
		hideErr:  true,
	}
}
