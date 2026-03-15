package vnstock

import (
	"errors"
	"fmt"
)

// ErrorCode represents a category of error returned by the library.
type ErrorCode string

const (
	// NetworkError indicates a network-level failure (DNS, TCP, TLS, timeout).
	NetworkError ErrorCode = "NETWORK_ERROR"
	// HTTPError indicates an HTTP response with a non-2xx status code.
	HTTPError ErrorCode = "HTTP_ERROR"
	// NotFound indicates the requested resource (symbol, index) was not found.
	NotFound ErrorCode = "NOT_FOUND"
	// NotSupported indicates the connector does not support the requested method.
	NotSupported ErrorCode = "NOT_SUPPORTED"
	// InvalidInput indicates invalid request parameters (e.g., start date after end date).
	InvalidInput ErrorCode = "INVALID_INPUT"
	// NoData indicates no data was available (empty listing, no financial data).
	NoData ErrorCode = "NO_DATA"
	// SerialiseError indicates a JSON/CSV serialisation or deserialisation failure.
	SerialiseError ErrorCode = "SERIALISE_ERROR"
	// ConfigError indicates an invalid configuration value.
	ConfigError ErrorCode = "CONFIG_ERROR"
)

// Error is the typed error returned by all library operations.
// It carries an error code, human-readable message, optional wrapped cause,
// and HTTP status code (when applicable).
type Error struct {
	Code       ErrorCode
	Message    string
	Cause      error
	StatusCode int // HTTP-specific; zero value when not applicable
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	if e.StatusCode != 0 {
		return fmt.Sprintf("%s: %s (HTTP %d)", e.Code, e.Message, e.StatusCode)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause, enabling errors.As and errors.Is traversal.
func (e *Error) Unwrap() error {
	return e.Cause
}

// AsError is a convenience wrapper around errors.As for *Error types.
func AsError(err error, target **Error) bool {
	if err == nil {
		return false
	}
	var e *Error
	if errors.As(err, &e) {
		*target = e
		return true
	}
	return false
}
