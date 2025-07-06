package model

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
)

// Application error codes
const (
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeInvalidInput       = "INVALID_INPUT"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeConflict           = "CONFLICT"
	ErrCodeInternal           = "INTERNAL_ERROR"
	ErrCodeStorage            = "STORAGE_ERROR"
	ErrCodeValidation         = "VALIDATION_ERROR"
	ErrCodeRateLimited        = "RATE_LIMITED"
	ErrCodeNotImplemented     = "NOT_IMPLEMENTED"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// AppError represents an application-specific error
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Details    any    `json:"details,omitempty"`
	Wrapped    error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("%s: %s (code: %s)", e.Message, e.Wrapped.Error(), e.Code)
	}
	return fmt.Sprintf("%s (code: %s)", e.Message, e.Code)
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Wrapped
}

// WithDetails adds additional context to the error
func (e *AppError) WithDetails(details any) *AppError {
	e.Details = details
	return e
}

// Wrap creates a new AppError that wraps another error
func (e *AppError) Wrap(err error) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		HTTPStatus: e.HTTPStatus,
		Wrapped:    err,
	}
}

// Predefined application errors
var (
	// 400 Bad Request
	ErrInvalidInput = &AppError{
		Code:       ErrCodeInvalidInput,
		Message:    "Invalid input provided",
		HTTPStatus: http.StatusBadRequest,
	}

	// 401 Unauthorized
	ErrUnauthorized = &AppError{
		Code:       ErrCodeUnauthorized,
		Message:    "Authentication required",
		HTTPStatus: http.StatusUnauthorized,
	}

	// 403 Forbidden
	ErrForbidden = &AppError{
		Code:       ErrCodeForbidden,
		Message:    "You don't have permission to access this resource",
		HTTPStatus: http.StatusForbidden,
	}

	// 404 Not Found
	ErrAssetNotFound = &AppError{
		Code:       ErrCodeNotFound,
		Message:    "Asset not found",
		HTTPStatus: http.StatusNotFound,
	}

	ErrUserNotFound = &AppError{
		Code:       ErrCodeNotFound,
		Message:    "User not found",
		HTTPStatus: http.StatusNotFound,
	}

	// 409 Conflict
	ErrAssetConflict = &AppError{
		Code:       ErrCodeConflict,
		Message:    "Asset already exists",
		HTTPStatus: http.StatusConflict,
	}

	// 422 Unprocessable Entity
	ErrValidationFailed = &AppError{
		Code:       ErrCodeValidation,
		Message:    "Validation failed",
		HTTPStatus: http.StatusUnprocessableEntity,
	}

	// 429 Too Many Requests
	ErrRateLimited = &AppError{
		Code:       ErrCodeRateLimited,
		Message:    "Too many requests",
		HTTPStatus: http.StatusTooManyRequests,
	}

	// 500 Internal Server Error
	ErrInternal = &AppError{
		Code:       ErrCodeInternal,
		Message:    "Internal server error",
		HTTPStatus: http.StatusInternalServerError,
	}

	// 501 Not Implemented
	ErrNotImplemented = &AppError{
		Code:       ErrCodeNotImplemented,
		Message:    "Feature not implemented",
		HTTPStatus: http.StatusNotImplemented,
	}

	// 503 Service Unavailable
	ErrServiceUnavailable = &AppError{
		Code:       ErrCodeServiceUnavailable,
		Message:    "Service temporarily unavailable",
		HTTPStatus: http.StatusServiceUnavailable,
	}
)

// Storage errors
var (
	ErrStorageUnavailable = &AppError{
		Code:       ErrCodeStorage,
		Message:    "Storage service unavailable",
		HTTPStatus: http.StatusServiceUnavailable,
	}

	ErrStorageTimeout = &AppError{
		Code:       ErrCodeStorage,
		Message:    "Storage operation timed out",
		HTTPStatus: http.StatusGatewayTimeout,
	}

	ErrStorageCorrupted = &AppError{
		Code:       ErrCodeStorage,
		Message:    "Data corruption detected",
		HTTPStatus: http.StatusInternalServerError,
	}
)

// Error helpers
func NewValidationError(field, message string) *AppError {
	return ErrValidationFailed.WithDetails(map[string]string{
		"field":   field,
		"message": message,
	})
}

func IsNotFoundError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == ErrCodeNotFound
	}
	return false
}

func IsConflictError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == ErrCodeConflict
	}
	return false
}

func WrapInternalError(err error) *AppError {
	return ErrInternal.Wrap(err)
}

func WrapStorageError(err error) *AppError {
	return ErrStorageUnavailable.Wrap(err)
}

// ConvertError converts standard errors to AppError
func ConvertError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	// Map common system errors to application errors
	switch {
	case errors.Is(err, os.ErrNotExist):
		return ErrAssetNotFound
	case errors.Is(err, os.ErrPermission):
		return ErrForbidden
	case errors.Is(err, context.DeadlineExceeded):
		return &AppError{
			Code:       "TIMEOUT",
			Message:    "Operation timed out",
			HTTPStatus: http.StatusGatewayTimeout,
			Wrapped:    err,
		}
	default:
		return WrapInternalError(err)
	}
}

// ErrorResponse defines the standard API error format
type ErrorResponse struct {
	Error      string `json:"error"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    any    `json:"details,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
	StatusCode int    `json:"-"`
}

// NewErrorResponse creates a standardized error response
func NewErrorResponse(err *AppError, requestID string) *ErrorResponse {
	return &ErrorResponse{
		Error:      err.Code,
		Code:       err.Code,
		Message:    err.Message,
		Details:    err.Details,
		RequestID:  requestID,
		StatusCode: err.HTTPStatus,
	}
}
