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
	"github.com/dda10/vnstock-go/connector/kbs"
)

func TestKBSNew(t *testing.T) {
	connector := kbs.New(&http.Client{}, slog.Default())
	if connector == nil {
		t.Fatal("expected non-nil connector")
	}
}

func TestKBSNewWithNilLogger(t *testing.T) {
	connector := kbs.New(&http.Client{}, nil)
	if connector == nil {
		t.Fatal("expected non-nil connector")
	}
}

// TestKBSQuoteHistory_ValidationErrors tests input validation through the public API.
func TestKBSQuoteHistory_ValidationErrors(t *testing.T) {
	connector := kbs.New(http.DefaultClient, slog.Default())
	ctx := context.Background()

	tests := []struct {
		name    string
		req     vnstock.QuoteHistoryRequest
		errCode vnstock.ErrorCode
	}{
		{
			name: "empty symbol",
			req: vnstock.QuoteHistoryRequest{
				Symbol:   "",
				Start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:      time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
				Interval: "1D",
			},
			errCode: vnstock.InvalidInput,
		},
		{
			name: "symbol too short",
			req: vnstock.QuoteHistoryRequest{
				Symbol:   "AB",
				Start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:      time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
				Interval: "1D",
			},
			errCode: vnstock.InvalidInput,
		},
		{
			name: "symbol too long",
			req: vnstock.QuoteHistoryRequest{
				Symbol:   "ABCDEFGHIJK",
				Start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:      time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
				Interval: "1D",
			},
			errCode: vnstock.InvalidInput,
		},
		{
			name: "symbol with special characters",
			req: vnstock.QuoteHistoryRequest{
				Symbol:   "VNM-X",
				Start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:      time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
				Interval: "1D",
			},
			errCode: vnstock.InvalidInput,
		},
		{
			name: "start after end",
			req: vnstock.QuoteHistoryRequest{
				Symbol:   "VNM",
				Start:    time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
				End:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Interval: "1D",
			},
			errCode: vnstock.InvalidInput,
		},
		{
			name: "start equals end",
			req: vnstock.QuoteHistoryRequest{
				Symbol:   "VNM",
				Start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Interval: "1D",
			},
			errCode: vnstock.InvalidInput,
		},
		{
			name: "invalid interval",
			req: vnstock.QuoteHistoryRequest{
				Symbol:   "VNM",
				Start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:      time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
				Interval: "invalid",
			},
			errCode: vnstock.InvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := connector.QuoteHistory(ctx, tt.req)
			if err == nil {
				t.Fatal("expected error")
			}
			var vErr *vnstock.Error
			if !vnstock.AsError(err, &vErr) {
				t.Fatalf("expected *vnstock.Error, got %T", err)
			}
			if vErr.Code != tt.errCode {
				t.Errorf("expected %v, got %v", tt.errCode, vErr.Code)
			}
		})
	}
}

// TestKBSQuoteHistory_HTTPErrors tests HTTP error handling via mock server.
func TestKBSQuoteHistory_HTTPErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"HTTP 400", http.StatusBadRequest},
		{"HTTP 500", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(map[string]string{"error": "test error"})
			}))
			defer server.Close()

			// Use a transport that redirects KBS requests to the mock server
			client := &http.Client{Transport: &redirectTransport{baseURL: server.URL}}
			connector := kbs.New(client, slog.Default())

			_, err := connector.QuoteHistory(context.Background(), vnstock.QuoteHistoryRequest{
				Symbol:   "VNM",
				Start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:      time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
				Interval: "1D",
			})

			if err == nil {
				t.Fatal("expected error")
			}
			var vErr *vnstock.Error
			if !vnstock.AsError(err, &vErr) {
				t.Fatalf("expected *vnstock.Error, got %T", err)
			}
			if vErr.Code != vnstock.HTTPError {
				t.Errorf("expected HTTPError, got %v", vErr.Code)
			}
			if vErr.StatusCode != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, vErr.StatusCode)
			}
		})
	}
}

// TestKBSQuoteHistory_Success tests successful quote retrieval.
func TestKBSQuoteHistory_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data_day": []map[string]interface{}{
				{"t": "2024-01-02", "o": 85000.0, "h": 86000.0, "l": 84500.0, "c": 85500.0, "v": 1234567},
				{"t": "2024-01-03", "o": 85500.0, "h": 87000.0, "l": 85000.0, "c": 86500.0, "v": 2345678},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &http.Client{Transport: &redirectTransport{baseURL: server.URL}}
	connector := kbs.New(client, slog.Default())

	quotes, err := connector.QuoteHistory(context.Background(), vnstock.QuoteHistoryRequest{
		Symbol:   "VNM",
		Start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:      time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
		Interval: "1D",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(quotes) != 2 {
		t.Errorf("expected 2 quotes, got %d", len(quotes))
	}
	if len(quotes) > 0 {
		if quotes[0].Symbol != "VNM" {
			t.Errorf("expected symbol VNM, got %s", quotes[0].Symbol)
		}
	}
}

// redirectTransport redirects all requests to a mock server URL.
type redirectTransport struct {
	baseURL string
}

func (rt *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	newURL := rt.baseURL + req.URL.Path
	if req.URL.RawQuery != "" {
		newURL += "?" + req.URL.RawQuery
	}
	newReq, err := http.NewRequestWithContext(req.Context(), req.Method, newURL, req.Body)
	if err != nil {
		return nil, err
	}
	newReq.Header = req.Header
	return http.DefaultTransport.RoundTrip(newReq)
}
