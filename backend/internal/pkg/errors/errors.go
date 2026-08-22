// Package errors defines the ImageForge error contract.
//
// Every API error maps to a stable, provider-independent code so the public
// API never leaks Sub2API/OpenAI/upstream internals.
package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode is a stable, machine-readable error identifier.
type ErrorCode string

const (
	ErrInvalidRequest       ErrorCode = "INVALID_REQUEST"
	ErrUnauthorized         ErrorCode = "UNAUTHORIZED"
	ErrForbidden            ErrorCode = "FORBIDDEN"
	ErrNotFound             ErrorCode = "NOT_FOUND"
	ErrConflict             ErrorCode = "CONFLICT"
	ErrImageTaskNotFound    ErrorCode = "IMAGE_TASK_NOT_FOUND"
	ErrImageTaskUnavailable ErrorCode = "IMAGE_TASK_UNAVAILABLE"
	ErrImageGeneration      ErrorCode = "IMAGE_GENERATION_FAILED"
	ErrImageEdit            ErrorCode = "IMAGE_EDIT_FAILED"
	ErrUpstreamTimeout      ErrorCode = "UPSTREAM_TIMEOUT"
	ErrUpstreamRateLimited  ErrorCode = "UPSTREAM_RATE_LIMITED"
	ErrModelUnavailable     ErrorCode = "MODEL_UNAVAILABLE"
	ErrAuthExpired          ErrorCode = "AUTHENTICATION_EXPIRED"
	ErrStorageFailed        ErrorCode = "STORAGE_FAILED"
	ErrQuotaExceeded        ErrorCode = "QUOTA_EXCEEDED"
	ErrInternal             ErrorCode = "INTERNAL_ERROR"
)

const (
	ErrorCodeInvalidRequest       = ErrInvalidRequest
	ErrorCodeUnauthorized         = ErrUnauthorized
	ErrorCodeForbidden            = ErrForbidden
	ErrorCodeNotFound             = ErrNotFound
	ErrorCodeConflict             = ErrConflict
	ErrorCodeImageTaskNotFound    = ErrImageTaskNotFound
	ErrorCodeImageTaskUnavailable = ErrImageTaskUnavailable
	ErrorCodeImageGeneration      = ErrImageGeneration
	ErrorCodeImageEdit            = ErrImageEdit
	ErrorCodeUpstreamTimeout      = ErrUpstreamTimeout
	ErrorCodeUpstreamRateLimited  = ErrUpstreamRateLimited
	ErrorCodeModelUnavailable     = ErrModelUnavailable
	ErrorCodeAuthExpired          = ErrAuthExpired
	ErrorCodeStorageFailed        = ErrStorageFailed
	ErrorCodeQuotaExceeded        = ErrQuotaExceeded
	ErrorCodeInternal             = ErrInternal
)

// Error is the application error type. HTTPStatus carries the response code;
// Code is the stable error code; Message is user-safe (never raw upstream).
type Error struct {
	Code       ErrorCode
	Message    string
	HTTPStatus int
	RequestID  string
	cause      error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.cause }

// Is allows errors.Is matching on code.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// New creates an *Error with the given code and default HTTP status.
func New(code ErrorCode, msg string) *Error {
	return &Error{Code: code, Message: msg, HTTPStatus: defaultHTTPStatus(code)}
}

// Wrap wraps an underlying cause while preserving the error code.
func Wrap(code ErrorCode, msg string, cause error) *Error {
	return &Error{Code: code, Message: msg, HTTPStatus: defaultHTTPStatus(code), cause: cause}
}

// WithRequestID attaches a request ID for correlation.
func (e *Error) WithRequestID(id string) *Error {
	e.RequestID = id
	return e
}

func defaultHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrInvalidRequest:
		return http.StatusBadRequest
	case ErrUnauthorized, ErrAuthExpired:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrNotFound:
		return http.StatusNotFound
	case ErrImageTaskNotFound:
		return http.StatusNotFound
	case ErrConflict:
		return http.StatusConflict
	case ErrUpstreamRateLimited:
		return http.StatusTooManyRequests
	case ErrUpstreamTimeout:
		return http.StatusGatewayTimeout
	case ErrQuotaExceeded:
		return http.StatusTooManyRequests
	case ErrInternal, ErrImageGeneration, ErrImageEdit, ErrStorageFailed, ErrModelUnavailable:
		return http.StatusInternalServerError
	case ErrImageTaskUnavailable:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
