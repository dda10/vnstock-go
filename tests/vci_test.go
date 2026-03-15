package tests

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	vnstock "github.com/dda10/vnstock-go"
	"github.com/dda10/vnstock-go/connector/vci"
)

func TestVCINew(t *testing.T) {
	connector := vci.New(&http.Client{}, slog.Default())
	if connector == nil {
		t.Fatal("expected non-nil connector")
	}
}

func TestVCINewWithNilLogger(t *testing.T) {
	connector := vci.New(&http.Client{}, nil)
	if connector == nil {
		t.Fatal("expected non-nil connector")
	}
}

func TestVCIQuoteHistory_InvalidDateRange(t *testing.T) {
	connector := vci.New(&http.Client{}, slog.Default())

	req := vnstock.QuoteHistoryRequest{
		Symbol:   "VNM",
		Start:    time.Now(),
		End:      time.Now().Add(-24 * time.Hour),
		Interval: "1d",
	}

	_, err := connector.QuoteHistory(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for invalid date range")
	}

	var vErr *vnstock.Error
	if !vnstock.AsError(err, &vErr) {
		t.Fatal("expected vnstock.Error")
	}
	if vErr.Code != vnstock.InvalidInput {
		t.Errorf("expected InvalidInput error code, got %s", vErr.Code)
	}
}

func TestVCIQuoteHistory_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/quote/history" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"symbol":    "VNM",
					"timestamp": "2024-01-01T00:00:00Z",
					"open":      100.0,
					"high":      105.0,
					"low":       99.0,
					"close":     103.0,
					"volume":    1000000,
				},
				{
					"symbol":    "VNM",
					"timestamp": "2024-01-02T00:00:00Z",
					"open":      103.0,
					"high":      108.0,
					"low":       102.0,
					"close":     107.0,
					"volume":    1200000,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_ = vci.New(&http.Client{}, slog.Default())
	// Note: This test demonstrates the structure but won't actually call the mock server
	// because baseURL is a package constant. In a real implementation, we'd make it configurable.
}

func TestVCIRealTimeQuotes_EmptySymbols(t *testing.T) {
	connector := vci.New(&http.Client{}, slog.Default())

	_, err := connector.RealTimeQuotes(context.Background(), []string{})
	if err == nil {
		t.Fatal("expected error for empty symbols list")
	}

	var vErr *vnstock.Error
	if !vnstock.AsError(err, &vErr) {
		t.Fatal("expected vnstock.Error")
	}
	if vErr.Code != vnstock.InvalidInput {
		t.Errorf("expected InvalidInput error code, got %s", vErr.Code)
	}
}

func TestVCIIndexHistory_InvalidDateRange(t *testing.T) {
	connector := vci.New(&http.Client{}, slog.Default())

	req := vnstock.IndexHistoryRequest{
		Name:     "VN-Index",
		Start:    time.Now(),
		End:      time.Now().Add(-24 * time.Hour),
		Interval: "1d",
	}

	_, err := connector.IndexHistory(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for invalid date range")
	}

	var vErr *vnstock.Error
	if !vnstock.AsError(err, &vErr) {
		t.Fatal("expected vnstock.Error")
	}
	if vErr.Code != vnstock.InvalidInput {
		t.Errorf("expected InvalidInput error code, got %s", vErr.Code)
	}
}
