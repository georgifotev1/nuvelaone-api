package errors

import (
	"errors"
	"fmt"
)

type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string { return e.Message }

func (e *AppError) Unwrap() error { return e.Err }

func (e *AppError) Cause() error { return e.Err }

var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrValidation   = errors.New("validation error")
	ErrInternal     = errors.New("internal error")
)

func NotFound(message string, err error) *AppError {
	return &AppError{Code: "NOT_FOUND", Message: message, Err: ErrNotFound}
}

func Conflict(message string) *AppError {
	return &AppError{Code: "CONFLICT", Message: message, Err: ErrConflict}
}

func Unauthorized(message string) *AppError {
	return &AppError{Code: "UNAUTHORIZED", Message: message, Err: ErrUnauthorized}
}

func Forbidden(message string) *AppError {
	return &AppError{Code: "FORBIDDEN", Message: message, Err: ErrForbidden}
}

func Validation(message string) *AppError {
	return &AppError{Code: "VALIDATION", Message: message, Err: ErrValidation}
}

func Internal(err error) *AppError {
	return &AppError{Code: "INTERNAL", Message: "an internal error occurred", Err: fmt.Errorf("%w: %w", ErrInternal, err)}
}
