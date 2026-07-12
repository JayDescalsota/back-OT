package errors

import "fmt"

type AppError struct {
	Code    string
	Message string
}

func (e *AppError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

func NotFound(resource string) *AppError {
	return &AppError{Code: "NOT_FOUND", Message: resource + " not found"}
}

func Validation(msg string) *AppError {
	return &AppError{Code: "VALIDATION_ERROR", Message: msg}
}

func Unauthorized(msg string) *AppError {
	return &AppError{Code: "UNAUTHORIZED", Message: msg}
}

func Forbidden(msg string) *AppError {
	return &AppError{Code: "FORBIDDEN", Message: msg}
}
