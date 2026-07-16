package response

import "fmt"

type Error struct {
	Code            string `json:"code"`
	Message         string `json:"message"`
	InternalMessage string `json:"internalMessage,omitempty"`
}

type SuccessResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *Error) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

func NotFound(resource string) *Error {
	return &Error{
		Code:    "NOT_FOUND",
		Message: resource + " not found",
	}
}

func Validation(msg string) *Error {
	return &Error{Code: "VALIDATION_ERROR", Message: msg}
}

func Unauthorized(msg string) *Error {
	return &Error{Code: "UNAUTHORIZED", Message: msg}
}

func Forbidden(msg string) *Error {
	return &Error{Code: "FORBIDDEN", Message: msg}
}

func Internal(msg string, internalMsg string) *Error {
	return &Error{
		Code:            "INTERNAL_ERROR",
		Message:         msg,
		InternalMessage: internalMsg,
	}
}

func NewError(code string, msg string, internalMsg string) *Error {
	return &Error{
		Code:            code,
		Message:         msg,
		InternalMessage: internalMsg,
	}
}

func Success(msg string) *SuccessResponse {
	return &SuccessResponse{
		Code:    "SUCCESS",
		Message: msg,
	}
}
