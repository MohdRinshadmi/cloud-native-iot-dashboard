// Package apperror defines a transport-agnostic error type used across the
// application and domain layers. Handlers map these to HTTP status codes,
// so business logic never imports net/http or Gin.
package apperror

import "fmt"

// Code is a stable, machine-readable error category.
type Code string

const (
	CodeInvalidInput Code = "INVALID_INPUT"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeForbidden    Code = "FORBIDDEN"
	CodeNotFound     Code = "NOT_FOUND"
	CodeConflict     Code = "CONFLICT"
	CodeInternal     Code = "INTERNAL"
	CodeUnavailable  Code = "UNAVAILABLE"
)

// Error is a structured application error carrying a category, a human message,
// and an optional wrapped cause.
type Error struct {
	Code    Code
	Message string
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap exposes the wrapped cause for errors.Is / errors.As.
func (e *Error) Unwrap() error { return e.cause }

// New builds an Error with the given code and message.
func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Wrap attaches a cause to a new Error.
func Wrap(code Code, message string, cause error) *Error {
	return &Error{Code: code, Message: message, cause: cause}
}

// Convenience constructors for the common cases.
func NotFound(msg string) *Error     { return New(CodeNotFound, msg) }
func InvalidInput(msg string) *Error { return New(CodeInvalidInput, msg) }
func Internal(cause error) *Error    { return Wrap(CodeInternal, "internal error", cause) }
