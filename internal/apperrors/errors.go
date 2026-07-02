package apperrors

import (
	"errors"
	"fmt"
)

// Sentinel-ошибки для единой обработки в middleware.
var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation")
	ErrInternal   = errors.New("internal")
)

// AppError связывает sentinel с HTTP-статусом и сообщением для клиента.
type AppError struct {
	Err     error
	Message string
	Status  int
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown error"
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NotFound(message string) *AppError {
	return &AppError{
		Err:     ErrNotFound,
		Message: message,
		Status:  404,
	}
}

func Validation(message string) *AppError {
	return &AppError{
		Err:     ErrValidation,
		Message: message,
		Status:  400,
	}
}

func Validationf(format string, args ...any) *AppError {
	return Validation(fmt.Sprintf(format, args...))
}

func Internal(message string) *AppError {
	if message == "" {
		message = "internal server error"
	}
	return &AppError{
		Err:     ErrInternal,
		Message: message,
		Status:  500,
	}
}

// FromSentinel преобразует sentinel или AppError в HTTP-ответ.
func FromSentinel(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return NotFound("resource not found")
	case errors.Is(err, ErrValidation):
		return Validation(err.Error())
	case errors.Is(err, ErrInternal):
		return Internal("")
	default:
		return Internal("internal server error")
	}
}
