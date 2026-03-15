// Package httpclient provides a factory for creating proxy-aware HTTP clients.
package httpclient

import (
	"errors"
	"net/http"
	"net/url"
	"time"
)

// ProxyError represents a proxy configuration error.
type ProxyError struct {
	Message string
	Cause   error
}

func (e *ProxyError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *ProxyError) Unwrap() error {
	return e.Cause
}

// IsProxyError checks if an error is a ProxyError.
func IsProxyError(err error) bool {
	var proxyErr *ProxyError
	return errors.As(err, &proxyErr)
}

// New creates an *http.Client configured with the given proxy URL and timeout.
// If proxyURL is empty, the client uses the default system transport (no proxy).
// If proxyURL is malformed, it returns a *ProxyError.
// The returned client uses http.Transport defaults for connection pooling.
func New(proxyURL string, timeout time.Duration) (*http.Client, error) {
	transport := &http.Transport{}

	if proxyURL != "" {
		parsed, err := url.Parse(proxyURL)
		if err != nil {
			return nil, &ProxyError{
				Message: "malformed proxy URL",
				Cause:   err,
			}
		}

		// Validate scheme is http or https
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return nil, &ProxyError{
				Message: "proxy URL must use http or https scheme",
			}
		}

		// Validate host is present
		if parsed.Host == "" {
			return nil, &ProxyError{
				Message: "proxy URL must include host",
			}
		}

		transport.Proxy = http.ProxyURL(parsed)
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}, nil
}
