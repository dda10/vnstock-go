package vnstock_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/dda10/vnstock-go"
	_ "github.com/dda10/vnstock-go/all" // Register all connectors
	"pgregory.net/rapid"
)

// Feature: vnstock-go-rewrite, Property 13: Invalid Config fields are rejected at construction
func TestInvalidConfigRejected(t *testing.T) {
	// Save and restore env vars
	oldConnector := os.Getenv("VNSTOCK_CONNECTOR")
	oldProxy := os.Getenv("VNSTOCK_PROXY_URL")
	defer func() {
		os.Setenv("VNSTOCK_CONNECTOR", oldConnector)
		os.Setenv("VNSTOCK_PROXY_URL", oldProxy)
	}()

	rapid.Check(t, func(t *rapid.T) {
		// Clear env vars for this test to ensure Config fields are used
		os.Unsetenv("VNSTOCK_CONNECTOR")
		os.Unsetenv("VNSTOCK_PROXY_URL")

		// Generate configs with at least one invalid field
		invalidType := rapid.IntRange(0, 2).Draw(t, "invalidType")

		cfg := vnstock.Config{}

		switch invalidType {
		case 0:
			// Negative timeout
			cfg.Timeout = time.Duration(rapid.Int64Range(-1000000000, -1).Draw(t, "negativeTimeout"))
			cfg.Connector = "VCI" // Valid connector to isolate the timeout error
		case 1:
			// Malformed proxy URL (missing port or invalid format)
			cfg.ProxyURL = "http://invalid:port:extra"
			cfg.Connector = "VCI"
			cfg.Timeout = 30 * time.Second
		case 2:
			// Empty connector (when env var is also not set)
			cfg.Connector = ""
			cfg.Timeout = 30 * time.Second
		}

		// Attempt to create client
		client, err := vnstock.New(cfg)

		// Should return an error
		if err == nil {
			t.Fatalf("expected error for invalid config, got nil")
		}

		// Should be nil client
		if client != nil {
			t.Fatalf("expected nil client for invalid config, got %v", client)
		}

		// Should be a ConfigError
		var libErr *vnstock.Error
		if !errors.As(err, &libErr) {
			t.Fatalf("expected *vnstock.Error, got %T", err)
		}
		if libErr.Code != vnstock.ConfigError {
			t.Fatalf("expected ConfigError, got %v", libErr.Code)
		}
	})
}

// Feature: vnstock-go-rewrite, Property 12: Unrecognised connector name is rejected at construction
func TestUnrecognisedConnectorRejected(t *testing.T) {
	// Save and restore env vars
	oldConnector := os.Getenv("VNSTOCK_CONNECTOR")
	oldProxy := os.Getenv("VNSTOCK_PROXY_URL")
	defer func() {
		os.Setenv("VNSTOCK_CONNECTOR", oldConnector)
		os.Setenv("VNSTOCK_PROXY_URL", oldProxy)
	}()

	rapid.Check(t, func(t *rapid.T) {
		// Clear env vars
		os.Unsetenv("VNSTOCK_CONNECTOR")
		os.Unsetenv("VNSTOCK_PROXY_URL")

		// Generate arbitrary strings that are NOT valid connector names
		validNames := []string{"VCI", "DNSE", "FMP", "Binance"}

		// Generate a random string that's not in the valid set
		invalidConnector := rapid.StringMatching(`[A-Za-z0-9_-]+`).
			Filter(func(s string) bool {
				for _, valid := range validNames {
					if s == valid {
						return false
					}
				}
				return s != "" // Also exclude empty string
			}).
			Draw(t, "invalidConnector")

		cfg := Config{
			Connector: invalidConnector,
			Timeout:   30 * time.Second,
		}

		// Attempt to create client
		client, err := vnstock.New(cfg)

		// Should return an error
		if err == nil {
			t.Fatalf("expected error for unrecognised connector %q, got nil", invalidConnector)
		}

		// Should be nil client
		if client != nil {
			t.Fatalf("expected nil client for unrecognised connector, got %v", client)
		}

		// Should be a ConfigError
		var libErr *vnstock.Error
		if !errors.As(err, &libErr) {
			t.Fatalf("expected *vnstock.Error, got %T", err)
		}
		if libErr.Code != vnstock.ConfigError {
			t.Fatalf("expected ConfigError, got %v", libErr.Code)
		}
	})
}
