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
	"github.com/dda10/vnstock-go/connector/vnd"
)

// TestVNDQuoteHistory_EmptySymbol validates Requirement 2.7:
// IF the symbol parameter is empty, THEN THE VND_Connector SHALL return an Error with Code InvalidInput.
func TestVNDQuoteHistory_EmptySymbol(t *testing.T) {
	connector := vnd.New(&http.Client{}, slog.Default())

	_, err := connector.QuoteHistory(context.Background(), vnstock.QuoteHistoryRequest{
		Symbol:   "",
		Start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:      time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
		Interval: "1D",
	})

	if err == nil {
		t.Fatal("expected error for empty symbol")
	}
	var vErr *vnstock.Error
	if !vnstock.AsError(err, &vErr) {
		t.Fatalf("expected *vnstock.Error, got %T", err)
	}
	if vErr.Code != vnstock.InvalidInput {
		t.Errorf("expected InvalidInput, got %v", vErr.Code)
	}
}

// TestVNDQuoteHistory_StartAfterEnd validates Requirement 2.8:
// IF the start date is after the end date, THEN THE VND_Connector SHALL return an Error with Code InvalidInput.
func TestVNDQuoteHistory_StartAfterEnd(t *testing.T) {
	connector := vnd.New(&http.Client{}, slog.Default())

	_, err := connector.QuoteHistory(context.Background(), vnstock.QuoteHistoryRequest{
		Symbol:   "VND",
		Start:    time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
		End:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Interval: "1D",
	})

	if err == nil {
		t.Fatal("expected error for start after end")
	}
	var vErr *vnstock.Error
	if !vnstock.AsError(err, &vErr) {
		t.Fatalf("expected *vnstock.Error, got %T", err)
	}
	if vErr.Code != vnstock.InvalidInput {
		t.Errorf("expected InvalidInput, got %v", vErr.Code)
	}
}

// TestVNDQuoteHistory_Success tests successful response parsing from a mock dchart API server.
func TestVNDQuoteHistory_Success(t *testing.T) {
	ts1 := int64(1704067200) // 2024-01-01 00:00:00 UTC
	ts2 := int64(1704153600) // 2024-01-02 00:00:00 UTC

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"s": "ok",
			"t": []int64{ts1, ts2},
			"o": []float64{15.5, 15.6},
			"h": []float64{15.8, 15.9},
			"l": []float64{15.3, 15.4},
			"c": []float64{15.7, 15.8},
			"v": []int64{1000000, 1200000},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &http.Client{Transport: &redirectTransport{baseURL: server.URL}}
	connector := vnd.New(client, slog.Default())

	quotes, err := connector.QuoteHistory(context.Background(), vnstock.QuoteHistoryRequest{
		Symbol:   "VND",
		Start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:      time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
		Interval: "1D",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(quotes) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(quotes))
	}

	// Validate first quote fields
	q := quotes[0]
	if q.Symbol != "VND" {
		t.Errorf("expected symbol VND, got %s", q.Symbol)
	}
	if q.Interval != "1D" {
		t.Errorf("expected interval 1D, got %s", q.Interval)
	}
	if q.Timestamp.Unix() != ts1 {
		t.Errorf("expected timestamp %d, got %d", ts1, q.Timestamp.Unix())
	}
	if q.Open != 15.5 {
		t.Errorf("expected open 15.5, got %f", q.Open)
	}
	if q.High != 15.8 {
		t.Errorf("expected high 15.8, got %f", q.High)
	}
	if q.Low != 15.3 {
		t.Errorf("expected low 15.3, got %f", q.Low)
	}
	if q.Close != 15.7 {
		t.Errorf("expected close 15.7, got %f", q.Close)
	}
	if q.Volume != 1000000 {
		t.Errorf("expected volume 1000000, got %d", q.Volume)
	}
}

// TestVNDListing_Success validates Requirements 8.1, 8.3:
// Listing with empty exchange returns all symbols mapped to []ListingRecord.
func TestVNDListing_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dchart/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("query") != "" {
			t.Errorf("expected empty query, got %s", r.URL.Query().Get("query"))
		}
		if r.URL.Query().Get("limit") != "30" {
			t.Errorf("expected limit=30, got %s", r.URL.Query().Get("limit"))
		}
		results := []map[string]string{
			{"symbol": "VND", "exchange": "HOSE", "description": "VNDirect Securities"},
			{"symbol": "VCI", "exchange": "HOSE", "description": "Viet Capital Securities"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}))
	defer server.Close()

	client := &http.Client{Transport: &redirectTransport{baseURL: server.URL}}
	connector := vnd.New(client, slog.Default())

	records, err := connector.Listing(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if records[0].Symbol != "VND" || records[0].Exchange != "HOSE" || records[0].CompanyName != "VNDirect Securities" {
		t.Errorf("unexpected first record: %+v", records[0])
	}
}

// TestVNDListing_ExchangeFilter validates Requirement 8.2:
// Listing with exchange filter passes the exchange parameter to the API.
func TestVNDListing_ExchangeFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("exchange") != "HNX" {
			t.Errorf("expected exchange=HNX, got %s", r.URL.Query().Get("exchange"))
		}
		results := []map[string]string{
			{"symbol": "SHB", "exchange": "HNX", "description": "Saigon-Hanoi Bank"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}))
	defer server.Close()

	client := &http.Client{Transport: &redirectTransport{baseURL: server.URL}}
	connector := vnd.New(client, slog.Default())

	records, err := connector.Listing(context.Background(), "HNX")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 1 || records[0].Exchange != "HNX" {
		t.Errorf("unexpected records: %+v", records)
	}
}

// TestVNDListing_NoData validates Requirement 8.4:
// Listing returns NoData error when the API returns an empty result set.
func TestVNDListing_NoData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := &http.Client{Transport: &redirectTransport{baseURL: server.URL}}
	connector := vnd.New(client, slog.Default())

	_, err := connector.Listing(context.Background(), "")
	if err == nil {
		t.Fatal("expected NoData error for empty results")
	}
	var vErr *vnstock.Error
	if !vnstock.AsError(err, &vErr) {
		t.Fatalf("expected *vnstock.Error, got %T", err)
	}
	if vErr.Code != vnstock.NoData {
		t.Errorf("expected NoData, got %v", vErr.Code)
	}
}
