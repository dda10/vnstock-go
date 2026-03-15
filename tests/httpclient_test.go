package tests

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/dda10/vnstock-go/internal/httpclient"
)

func TestHTTPClient_NoProxy(t *testing.T) {
	timeout := 30 * time.Second
	client, err := httpclient.New("", timeout)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.Timeout != timeout {
		t.Errorf("expected timeout %v, got %v", timeout, client.Timeout)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}
	if transport.Proxy != nil {
		t.Error("expected nil proxy function for no-proxy path")
	}
}

func TestHTTPClient_ValidHTTPProxy(t *testing.T) {
	client, err := httpclient.New("http://proxy.example.com:8080", 15*time.Second)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}
	if transport.Proxy == nil {
		t.Error("expected proxy function to be set")
	}
}

func TestHTTPClient_ValidHTTPSProxy(t *testing.T) {
	client, err := httpclient.New("https://secure-proxy.example.com:443", 20*time.Second)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}
	if transport.Proxy == nil {
		t.Error("expected proxy function to be set")
	}
}

func TestHTTPClient_MalformedProxyURL(t *testing.T) {
	testCases := []struct {
		name     string
		proxyURL string
		wantMsg  string
	}{
		{"invalid URL characters", "http://proxy.example.com:8080\x00invalid", "malformed proxy URL"},
		{"missing scheme", "://proxy.example.com:8080", "malformed proxy URL"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := httpclient.New(tc.proxyURL, 30*time.Second)
			if client != nil {
				t.Error("expected nil client for malformed URL")
			}
			if err == nil {
				t.Fatal("expected error for malformed URL")
			}

			var proxyErr *httpclient.ProxyError
			if !errors.As(err, &proxyErr) {
				t.Fatalf("expected *ProxyError, got %T", err)
			}
			if proxyErr.Message != tc.wantMsg {
				t.Errorf("expected message %q, got %q", tc.wantMsg, proxyErr.Message)
			}
		})
	}
}

func TestHTTPClient_InvalidProxyScheme(t *testing.T) {
	testCases := []string{
		"ftp://proxy.example.com:21",
		"socks5://proxy.example.com:1080",
		"file:///etc/proxy",
	}

	for _, proxyURL := range testCases {
		t.Run(proxyURL, func(t *testing.T) {
			client, err := httpclient.New(proxyURL, 30*time.Second)
			if client != nil {
				t.Error("expected nil client for invalid scheme")
			}
			if err == nil {
				t.Fatal("expected error for invalid scheme")
			}

			var proxyErr *httpclient.ProxyError
			if !errors.As(err, &proxyErr) {
				t.Fatalf("expected *ProxyError, got %T", err)
			}
			if proxyErr.Message != "proxy URL must use http or https scheme" {
				t.Errorf("unexpected error message: %s", proxyErr.Message)
			}
		})
	}
}

func TestHTTPClient_ProxyURLMissingHost(t *testing.T) {
	testCases := []string{"http://", "https://"}

	for _, proxyURL := range testCases {
		t.Run(proxyURL, func(t *testing.T) {
			client, err := httpclient.New(proxyURL, 30*time.Second)
			if client != nil {
				t.Error("expected nil client for missing host")
			}
			if err == nil {
				t.Fatal("expected error for missing host")
			}

			var proxyErr *httpclient.ProxyError
			if !errors.As(err, &proxyErr) {
				t.Fatalf("expected *ProxyError, got %T", err)
			}
			if proxyErr.Message != "proxy URL must include host" {
				t.Errorf("unexpected error message: %s", proxyErr.Message)
			}
		})
	}
}

func TestHTTPClient_ZeroTimeout(t *testing.T) {
	client, err := httpclient.New("", 0)
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

func TestHTTPClient_ProxyWithCredentials(t *testing.T) {
	client, err := httpclient.New("http://user:password@proxy.example.com:8080", 30*time.Second)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}
	if transport.Proxy == nil {
		t.Error("expected proxy function to be set")
	}
}
