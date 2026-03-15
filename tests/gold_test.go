package tests

import (
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"

	vnstock "github.com/dda10/vnstock-go"
	"github.com/dda10/vnstock-go/connector/gold"
)

func newGoldConnector() *gold.Connector {
	client := &http.Client{Timeout: 30 * time.Second}
	c := gold.New(client, slog.Default())
	return c.(*gold.Connector)
}

func TestGoldConnector_SJCPrices(t *testing.T) {
	connector := newGoldConnector()
	ctx := context.Background()
	req := vnstock.GoldPriceRequest{Date: time.Now(), Source: "SJC"}

	prices, err := connector.GoldPrice(ctx, req)
	if err != nil {
		t.Skip("Skipping test - SJC API unavailable")
	}
	if len(prices) == 0 {
		t.Fatal("Expected at least one gold price record from SJC")
	}

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
		t.Errorf("Expected Source 'SJC', got '%s'", first.Source)
	}
}

func TestGoldConnector_BTMCPrices(t *testing.T) {
	connector := newGoldConnector()
	ctx := context.Background()
	req := vnstock.GoldPriceRequest{Source: "BTMC"}

	prices, err := connector.GoldPrice(ctx, req)
	if err != nil {
		t.Skip("Skipping test - BTMC API unavailable")
	}
	if len(prices) == 0 {
		t.Fatal("Expected at least one gold price record from BTMC")
	}

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
		t.Errorf("Expected Source 'BTMC', got '%s'", first.Source)
	}
}

func TestGoldConnector_AllSources(t *testing.T) {
	connector := newGoldConnector()
	ctx := context.Background()
	req := vnstock.GoldPriceRequest{Date: time.Now()}

	prices, err := connector.GoldPrice(ctx, req)
	if err != nil {
		t.Skip("Skipping test - APIs unavailable")
	}
	if len(prices) == 0 {
		t.Fatal("Expected at least one gold price record from any source")
	}

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
	connector := newGoldConnector()
	ctx := context.Background()
	historicalDate := time.Now().AddDate(0, 0, -2)
	req := vnstock.GoldPriceRequest{Date: historicalDate, Source: "SJC"}

	prices, err := connector.GoldPrice(ctx, req)
	if err != nil {
		t.Skip("Skipping test - SJC API unavailable")
	}
	if len(prices) == 0 {
		t.Fatal("Expected at least one historical gold price record")
	}
}

func TestGoldConnector_UnsupportedMethods(t *testing.T) {
	client := &http.Client{Timeout: 30 * time.Second}
	connector := gold.New(client, slog.Default())
	ctx := context.Background()

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
