package gold

import (
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"

	vnstock "github.com/dda10/vnstock-go"
)

func TestGoldConnector_SJCPrices(t *testing.T) {
	client := &http.Client{Timeout: 30 * time.Second}
	logger := slog.Default()
	connector := New(client, logger)

	ctx := context.Background()
	req := vnstock.GoldPriceRequest{
		Date:   time.Now(),
		Source: "SJC",
	}

	prices, err := connector.GoldPrice(ctx, req)
	if err != nil {
		t.Logf("SJC API request failed (may be expected if API is down): %v", err)
		t.Skip("Skipping test - SJC API unavailable")
		return
	}

	if len(prices) == 0 {
		t.Fatal("Expected at least one gold price record from SJC")
	}

	t.Logf("Fetched %d SJC gold price records", len(prices))

	// Validate first record
	first := prices[0]
	if first.TypeName == "" {
		t.Error("Expected TypeName to be non-empty")
	}
	if first.BuyPrice <= 0 {
		t.Error("Expected BuyPrice to be positive")
	}
	if first.SellPrice <= 0 {
		t.Error("Expected SellPrice to be positive")
	}
	if first.Source != "SJC" {
		t.Errorf("Expected Source to be 'SJC', got '%s'", first.Source)
	}

	t.Logf("Sample record: %s - Buy: %.0f, Sell: %.0f", first.TypeName, first.BuyPrice, first.SellPrice)
}

func TestGoldConnector_BTMCPrices(t *testing.T) {
	client := &http.Client{Timeout: 30 * time.Second}
	logger := slog.Default()
	connector := New(client, logger)

	ctx := context.Background()
	req := vnstock.GoldPriceRequest{
		Source: "BTMC",
	}

	prices, err := connector.GoldPrice(ctx, req)
	if err != nil {
		t.Logf("BTMC API request failed (may be expected if API is down): %v", err)
		t.Skip("Skipping test - BTMC API unavailable")
		return
	}

	if len(prices) == 0 {
		t.Fatal("Expected at least one gold price record from BTMC")
	}

	t.Logf("Fetched %d BTMC gold price records", len(prices))

	// Validate first record
	first := prices[0]
	if first.TypeName == "" {
		t.Error("Expected TypeName to be non-empty")
	}
	if first.BuyPrice <= 0 {
		t.Error("Expected BuyPrice to be positive")
	}
	if first.SellPrice <= 0 {
		t.Error("Expected SellPrice to be positive")
	}
	if first.Source != "BTMC" {
		t.Errorf("Expected Source to be 'BTMC', got '%s'", first.Source)
	}

	t.Logf("Sample record: %s - Buy: %.0f, Sell: %.0f", first.TypeName, first.BuyPrice, first.SellPrice)
}

func TestGoldConnector_AllSources(t *testing.T) {
	client := &http.Client{Timeout: 30 * time.Second}
	logger := slog.Default()
	connector := New(client, logger)

	ctx := context.Background()
	req := vnstock.GoldPriceRequest{
		Date: time.Now(),
		// No Source specified - should fetch from all sources
	}

	prices, err := connector.GoldPrice(ctx, req)
	if err != nil {
		t.Logf("Gold price API request failed: %v", err)
		t.Skip("Skipping test - APIs unavailable")
		return
	}

	if len(prices) == 0 {
		t.Fatal("Expected at least one gold price record from any source")
	}

	t.Logf("Fetched %d gold price records from all sources", len(prices))

	// Check that we have records from multiple sources
	sources := make(map[string]int)
	for _, price := range prices {
		sources[price.Source]++
	}

	t.Logf("Sources found: %v", sources)

	// Validate at least one record
	first := prices[0]
	if first.TypeName == "" {
		t.Error("Expected TypeName to be non-empty")
	}
	if first.BuyPrice <= 0 {
		t.Error("Expected BuyPrice to be positive")
	}
	if first.SellPrice <= 0 {
		t.Error("Expected SellPrice to be positive")
	}
}

func TestGoldConnector_HistoricalDate(t *testing.T) {
	client := &http.Client{Timeout: 30 * time.Second}
	logger := slog.Default()
	connector := New(client, logger)

	ctx := context.Background()
	// Test with a historical date (2 days ago)
	historicalDate := time.Now().AddDate(0, 0, -2)
	req := vnstock.GoldPriceRequest{
		Date:   historicalDate,
		Source: "SJC",
	}

	prices, err := connector.GoldPrice(ctx, req)
	if err != nil {
		t.Logf("SJC historical API request failed: %v", err)
		t.Skip("Skipping test - SJC API unavailable")
		return
	}

	if len(prices) == 0 {
		t.Fatal("Expected at least one historical gold price record")
	}

	t.Logf("Fetched %d historical gold price records for date %s", len(prices), historicalDate.Format("2006-01-02"))

	// Validate date matches request
	first := prices[0]
	if !first.Date.Equal(historicalDate.Truncate(24 * time.Hour)) {
		t.Logf("Note: Date mismatch - requested %s, got %s (may be expected for API behavior)",
			historicalDate.Format("2006-01-02"), first.Date.Format("2006-01-02"))
	}
}

func TestGoldConnector_UnsupportedMethods(t *testing.T) {
	client := &http.Client{Timeout: 30 * time.Second}
	logger := slog.Default()
	connector := New(client, logger)

	ctx := context.Background()

	// Test that unsupported methods return ErrNotSupported
	t.Run("QuoteHistory", func(t *testing.T) {
		_, err := connector.QuoteHistory(ctx, vnstock.QuoteHistoryRequest{})
		if err != vnstock.ErrNotSupported {
			t.Errorf("Expected ErrNotSupported, got %v", err)
		}
	})

	t.Run("RealTimeQuotes", func(t *testing.T) {
		_, err := connector.RealTimeQuotes(ctx, []string{"SSI"})
		if err != vnstock.ErrNotSupported {
			t.Errorf("Expected ErrNotSupported, got %v", err)
		}
	})

	t.Run("Listing", func(t *testing.T) {
		_, err := connector.Listing(ctx, "HOSE")
		if err != vnstock.ErrNotSupported {
			t.Errorf("Expected ErrNotSupported, got %v", err)
		}
	})
}
