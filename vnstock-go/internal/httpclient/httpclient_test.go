// Package httpclient provides tests for the HTTP client factory.
package httpclient

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

// TestNew_NoProxy verifies that when no proxy URL is provided,
// the client uses the default transport with no proxy configured.
// Validates: Requirements 8.4
func TestNew_NoProxy(t *testing.T) {
	timeout := 30 * time.Second
	client, err := New("", timeout)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.Timeout != timeout {
		t.Errorf("expected timeout %v, got %v", timeout, client.Timeout)
	}

	// Verify transport is set and has no proxy
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}
	if transport.Proxy != nil {
		t.Error("expected nil proxy function for no-proxy path")
	}
}

// TestNew_ValidHTTPProxy verifies that a valid HTTP proxy URL
// configures the transport correctly.
// Validates: Requirements 8.1, 8.3
func TestNew_ValidHTTPProxy(t *testing.T) {
	proxyURL := "http://proxy.example.com:8080"
	timeout := 15 * time.Second

	client, err := New(proxyURL, timeout)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.Timeout != timeout {
		t.Errorf("expected timeout %v, got %v", timeout, client.Timeout)
	}

	// Verify transport has proxy configured
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}
	if transport.Proxy == nil {
		t.Error("expected proxy function to be set")
	}
}

// TestNew_ValidHTTPSProxy verifies that a valid HTTPS proxy URL
// configures the transport correctly.
// Validates: Requirements 8.1, 8.3
func TestNew_ValidHTTPSProxy(t *testing.T) {
	proxyURL := "https://secure-proxy.example.com:443"
	timeout := 20 * time.Second

	client, err := New(proxyURL, timeout)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	// Verify transport has proxy configured
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}
	if transport.Proxy == nil {
		t.Error("expected proxy function to be set")
	}
}

// TestNew_MalformedProxyURL verifies that a malformed proxy URL
// returns a ProxyError.
// Validates: Requirements 8.3
func TestNew_MalformedProxyURL(t *testing.T) {
	testCases := []struct {
		name     string
		proxyURL string
		wantMsg  string
	}{
		{
			name:     "invalid URL characters",
			proxyURL: "http://proxy.example.com:8080\x00invalid",
			wantMsg:  "malformed proxy URL",
		},
		{
			name:     "missing scheme",
			proxyURL: "://proxy.example.com:8080",
			wantMsg:  "malformed proxy URL",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := New(tc.proxyURL, 30*time.Second)
			if client != nil {
				t.Error("expected nil client for malformed URL")
			}
			if err == nil {
				t.Fatal("expected error for malformed URL")
			}

			var proxyErr *ProxyError
			if !errors.As(err, &proxyErr) {
				t.Fatalf("expected *ProxyError, got %T", err)
			}
			if proxyErr.Message != tc.wantMsg {
				t.Errorf("expected message %q, got %q", tc.wantMsg, proxyErr.Message)
			}
		})
	}
}

// TestNew_InvalidProxyScheme verifies that proxy URLs with invalid schemes
// (not http or https) return a ProxyError.
// Validates: Requirements 8.3
func TestNew_InvalidProxyScheme(t *testing.T) {
	testCases := []struct {
		name     string
		proxyURL string
	}{
		{
			name:     "ftp scheme",
			proxyURL: "ftp://proxy.example.com:21",
		},
		{
			name:     "socks scheme",
			proxyURL: "socks5://proxy.example.com:1080",
		},
		{
			name:     "file scheme",
			proxyURL: "file:///etc/proxy",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := New(tc.proxyURL, 30*time.Second)
			if client != nil {
				t.Error("expected nil client for invalid scheme")
			}
			if err == nil {
				t.Fatal("expected error for invalid scheme")
			}

			var proxyErr *ProxyError
			if !errors.As(err, &proxyErr) {
				t.Fatalf("expected *ProxyError, got %T", err)
			}
			if proxyErr.Message != "proxy URL must use http or https scheme" {
				t.Errorf("unexpected error message: %s", proxyErr.Message)
			}
		})
	}
}

// TestNew_ProxyURLMissingHost verifies that proxy URLs without a host
// return a ProxyError.
// Validates: Requirements 8.3
func TestNew_ProxyURLMissingHost(t *testing.T) {
	testCases := []struct {
		name     string
		proxyURL string
	}{
		{
			name:     "http scheme only",
			proxyURL: "http://",
		},
		{
			name:     "https scheme only",
			proxyURL: "https://",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := New(tc.proxyURL, 30*time.Second)
			if client != nil {
				t.Error("expected nil client for missing host")
			}
			if err == nil {
				t.Fatal("expected error for missing host")
			}

			var proxyErr *ProxyError
			if !errors.As(err, &proxyErr) {
				t.Fatalf("expected *ProxyError, got %T", err)
			}
			if proxyErr.Message != "proxy URL must include host" {
				t.Errorf("unexpected error message: %s", proxyErr.Message)
			}
		})
	}
}

// TestNew_ZeroTimeout verifies that a zero timeout is accepted
// (no timeout on the client).
// Validates: Requirements 8.1
func TestNew_ZeroTimeout(t *testing.T) {
	client, err := New("", 0)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.Timeout != 0 {
		t.Errorf("expected zero timeout, got %v", client.Timeout)
	}
}

// TestNew_ProxyWithCredentials verifies that proxy URLs with
// embedded credentials are accepted.
// Validates: Requirements 8.1, 8.3
func TestNew_ProxyWithCredentials(t *testing.T) {
	proxyURL := "http://user:password@proxy.example.com:8080"
	timeout := 30 * time.Second

	client, err := New(proxyURL, timeout)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	// Verify transport has proxy configured
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}
	if transport.Proxy == nil {
		t.Error("expected proxy function to be set")
	}
}
