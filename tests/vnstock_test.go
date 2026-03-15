package tests

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/dda10/vnstock-go"
	_ "github.com/dda10/vnstock-go/all"
	"pgregory.net/rapid"
)

// Feature: vnstock-go-rewrite, Property 13: Invalid Config fields are rejected at construction
func TestInvalidConfigRejected(t *testing.T) {
	oldConnector := os.Getenv("VNSTOCK_CONNECTOR")
	oldProxy := os.Getenv("VNSTOCK_PROXY_URL")
	defer func() {
		os.Setenv("VNSTOCK_CONNECTOR", oldConnector)
		os.Setenv("VNSTOCK_PROXY_URL", oldProxy)
	}()

	rapid.Check(t, func(t *rapid.T) {
		os.Unsetenv("VNSTOCK_CONNECTOR")
		os.Unsetenv("VNSTOCK_PROXY_URL")

		invalidType := rapid.IntRange(0, 2).Draw(t, "invalidType")
		cfg := vnstock.Config{}

		switch invalidType {
		case 0:
			cfg.Timeout = time.Duration(rapid.Int64Range(-1000000000, -1).Draw(t, "negativeTimeout"))
			cfg.Connector = "VCI"
		case 1:
			cfg.ProxyURL = "http://invalid:port:extra"
			cfg.Connector = "VCI"
			cfg.Timeout = 30 * time.Second
		case 2:
			cfg.Connector = ""
			cfg.Timeout = 30 * time.Second
		}

		client, err := vnstock.New(cfg)
		if err == nil {
			t.Fatalf("expected error for invalid config, got nil")
		}
		if client != nil {
			t.Fatalf("expected nil client for invalid config, got %v", client)
		}

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
	oldConnector := os.Getenv("VNSTOCK_CONNECTOR")
	oldProxy := os.Getenv("VNSTOCK_PROXY_URL")
	defer func() {
		os.Setenv("VNSTOCK_CONNECTOR", oldConnector)
		os.Setenv("VNSTOCK_PROXY_URL", oldProxy)
	}()

	rapid.Check(t, func(t *rapid.T) {
		os.Unsetenv("VNSTOCK_CONNECTOR")
		os.Unsetenv("VNSTOCK_PROXY_URL")

		validNames := []string{"VCI", "DNSE", "FMP", "Binance"}
		invalidConnector := rapid.StringMatching(`[A-Za-z0-9_-]+`).
			Filter(func(s string) bool {
				for _, valid := range validNames {
					if s == valid {
						return false
					}
				}
				return s != ""
			}).
			Draw(t, "invalidConnector")

		cfg := vnstock.Config{
			Connector: invalidConnector,
			Timeout:   30 * time.Second,
		}

		client, err := vnstock.New(cfg)
		if err == nil {
			t.Fatalf("expected error for unrecognised connector %q, got nil", invalidConnector)
		}
		if client != nil {
			t.Fatalf("expected nil client for unrecognised connector, got %v", client)
		}

		var libErr *vnstock.Error
		if !errors.As(err, &libErr) {
			t.Fatalf("expected *vnstock.Error, got %T", err)
		}
		if libErr.Code != vnstock.ConfigError {
			t.Fatalf("expected ConfigError, got %v", libErr.Code)
		}
	})
}
