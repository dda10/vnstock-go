package kbs

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	vnstock "github.com/dda10/vnstock-go"
)

// TestValidateSymbol tests the validateSymbol helper function.
func TestValidateSymbol(t *testing.T) {
	tests := []struct {
		name    string
		symbol  string
		wantErr bool
		errCode vnstock.ErrorCode
	}{
		{
			name:    "valid symbol uppercase",
			symbol:  "VNM",
			wantErr: false,
		},
		{
			name:    "valid symbol lowercase",
			symbol:  "vnm",
			wantErr: false,
		},
		{
			name:    "valid symbol with numbers",
			symbol:  "VN30",
			wantErr: false,
		},
		{
			name:    "valid symbol max length",
			symbol:  "ABCDEFGHIJ",
			wantErr: false,
		},
		{
			name:    "empty symbol",
			symbol:  "",
			wantErr: true,
			errCode: vnstock.InvalidInput,
		},
		{
			name:    "symbol too short",
			symbol:  "AB",
			wantErr: true,
			errCode: vnstock.InvalidInput,
		},
		{
			name:    "symbol too long",
			symbol:  "ABCDEFGHIJK",
			wantErr: true,
			errCode: vnstock.InvalidInput,
		},
		{
			name:    "symbol with special characters",
			symbol:  "VNM-X",
			wantErr: true,
			errCode: vnstock.InvalidInput,
		},
		{
			name:    "symbol with spaces",
			symbol:  "VN M",
			wantErr: true,
			errCode: vnstock.InvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSymbol(tt.symbol)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSymbol() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				var vErr *vnstock.Error
				if !vnstock.AsError(err, &vErr) {
					t.Errorf("validateSymbol() error is not *vnstock.Error: %v", err)
					return
				}
				if vErr.Code != tt.errCode {
					t.Errorf("validateSymbol() error code = %v, want %v", vErr.Code, tt.errCode)
				}
			}
		})
	}
}

// TestValidateDateRange tests the validateDateRange helper function.
func TestValidateDateRange(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	tests := []struct {
		name    string
		start   time.Time
		end     time.Time
		wantErr bool
		errCode vnstock.ErrorCode
	}{
		{
			name:    "valid range",
			start:   yesterday,
			end:     now,
			wantErr: false,
		},
		{
			name:    "start after end",
			start:   tomorrow,
			end:     now,
			wantErr: true,
			errCode: vnstock.InvalidInput,
		},
		{
			name:    "start equals end",
			start:   now,
			end:     now,
			wantErr: true,
			errCode: vnstock.InvalidInput,
		},
		{
			name:    "valid long range",
			start:   now.AddDate(-1, 0, 0),
			end:     now,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDateRange(tt.start, tt.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDateRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				var vErr *vnstock.Error
				if !vnstock.AsError(err, &vErr) {
					t.Errorf("validateDateRange() error is not *vnstock.Error: %v", err)
					return
				}
				if vErr.Code != tt.errCode {
					t.Errorf("validateDateRange() error code = %v, want %v", vErr.Code, tt.errCode)
				}
			}
		})
	}
}

// TestDoRequest tests the doRequest helper function with mocked HTTP responses.
func TestDoRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		endpoint       string
		payload        interface{}
		mockStatusCode int
		mockResponse   interface{}
		wantErr        bool
		errCode        vnstock.ErrorCode
	}{
		{
			name:           "successful GET request",
			method:         http.MethodGet,
			endpoint:       "/test",
			payload:        nil,
			mockStatusCode: http.StatusOK,
			mockResponse:   map[string]string{"status": "ok"},
			wantErr:        false,
		},
		{
			name:           "successful POST request with payload",
			method:         http.MethodPost,
			endpoint:       "/test",
			payload:        map[string]string{"key": "value"},
			mockStatusCode: http.StatusOK,
			mockResponse:   map[string]string{"status": "ok"},
			wantErr:        false,
		},
		{
			name:           "HTTP 400 error",
			method:         http.MethodGet,
			endpoint:       "/test",
			payload:        nil,
			mockStatusCode: http.StatusBadRequest,
			mockResponse:   map[string]string{"error": "bad request"},
			wantErr:        true,
			errCode:        vnstock.HTTPError,
		},
		{
			name:           "HTTP 500 error",
			method:         http.MethodGet,
			endpoint:       "/test",
			payload:        nil,
			mockStatusCode: http.StatusInternalServerError,
			mockResponse:   map[string]string{"error": "internal server error"},
			wantErr:        true,
			errCode:        vnstock.HTTPError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify headers
				acceptHeader := r.Header.Get("Accept")
				if acceptHeader == "" {
					t.Error("Expected Accept header to be set")
				}
				if tt.payload != nil && r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type header to be application/json, got %s", r.Header.Get("Content-Type"))
				}

				w.WriteHeader(tt.mockStatusCode)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Create connector with mock server URL
			logger := slog.Default()
			c := &Connector{
				client: server.Client(),
				logger: logger,
			}

			// Use the server URL directly (doRequest now accepts full URLs)
			ctx := context.Background()
			resp, err := c.doRequest(ctx, tt.method, server.URL+tt.endpoint, tt.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("doRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				var vErr *vnstock.Error
				if !vnstock.AsError(err, &vErr) {
					t.Errorf("doRequest() error is not *vnstock.Error: %v", err)
					return
				}
				if vErr.Code != tt.errCode {
					t.Errorf("doRequest() error code = %v, want %v", vErr.Code, tt.errCode)
				}
				if tt.errCode == vnstock.HTTPError && vErr.StatusCode != tt.mockStatusCode {
					t.Errorf("doRequest() status code = %v, want %v", vErr.StatusCode, tt.mockStatusCode)
				}
			} else {
				if resp == nil {
					t.Error("doRequest() returned nil response without error")
					return
				}
				defer resp.Body.Close()
				if resp.StatusCode != tt.mockStatusCode {
					t.Errorf("doRequest() status code = %v, want %v", resp.StatusCode, tt.mockStatusCode)
				}
			}
		})
	}
}

// TestDoRequestInvalidPayload tests doRequest with invalid JSON payload.
func TestDoRequestInvalidPayload(t *testing.T) {
	logger := slog.Default()
	c := &Connector{
		client: http.DefaultClient,
		logger: logger,
	}

	// Create a payload that cannot be marshaled to JSON
	invalidPayload := make(chan int)

	ctx := context.Background()
	_, err := c.doRequest(ctx, http.MethodPost, "/test", invalidPayload)

	if err == nil {
		t.Error("doRequest() expected error for invalid payload, got nil")
		return
	}

	var vErr *vnstock.Error
	if !vnstock.AsError(err, &vErr) {
		t.Errorf("doRequest() error is not *vnstock.Error: %v", err)
		return
	}

	if vErr.Code != vnstock.SerialiseError {
		t.Errorf("doRequest() error code = %v, want %v", vErr.Code, vnstock.SerialiseError)
	}
}

// TestLogRequest tests the logRequest helper function.
func TestLogRequest(t *testing.T) {
	// Create a custom logger that captures log output
	var logOutput []string
	handler := slog.NewTextHandler(&testWriter{output: &logOutput}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)

	c := &Connector{
		client: http.DefaultClient,
		logger: logger,
	}

	// Call logRequest
	c.logRequest(http.MethodGet, "https://api.kbs.com.vn/test", http.StatusOK, 100*time.Millisecond)

	// Verify that log was written
	if len(logOutput) == 0 {
		t.Error("logRequest() did not produce any log output")
	}
}

// testWriter is a simple io.Writer that captures output to a string slice.
type testWriter struct {
	output *[]string
}

func (w *testWriter) Write(p []byte) (n int, err error) {
	*w.output = append(*w.output, string(p))
	return len(p), nil
}

// TestQuoteHistory tests the QuoteHistory method with mocked HTTP responses.
func TestQuoteHistory(t *testing.T) {
	tests := []struct {
		name           string
		req            vnstock.QuoteHistoryRequest
		mockStatusCode int
		mockResponse   map[string]interface{}
		wantErr        bool
		errCode        vnstock.ErrorCode
		wantQuoteCount int
		validateQuote  func(t *testing.T, quote vnstock.Quote)
	}{
		{
			name: "successful daily quotes",
			req: vnstock.QuoteHistoryRequest{
				Symbol:   "VNM",
				Start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:      time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
				Interval: "1D",
			},
			mockStatusCode: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data_day": []map[string]interface{}{
					{
						"t": "2024-01-02",
						"o": 85000.0,
						"h": 86000.0,
						"l": 84500.0,
						"c": 85500.0,
						"v": 1234567,
					},
					{
						"t": "2024-01-03",
						"o": 85500.0,
						"h": 87000.0,
						"l": 85000.0,
						"c": 86500.0,
						"v": 2345678,
					},
				},
			},
			wantErr:        false,
			wantQuoteCount: 2,
			validateQuote: func(t *testing.T, quote vnstock.Quote) {
				if quote.Symbol != "VNM" {
					t.Errorf("Quote.Symbol = %v, want VNM", quote.Symbol)
				}
				if quote.Interval != "1D" {
					t.Errorf("Quote.Interval = %v, want 1D", quote.Interval)
				}
				if quote.Open == 0 || quote.High == 0 || quote.Low == 0 || quote.Close == 0 {
					t.Error("Quote OHLC values should not be zero")
				}
				if quote.Volume == 0 {
					t.Error("Quote.Volume should not be zero")
				}
			},
		},
		{
			name: "successful intraday 1 minute quotes",
			req: vnstock.QuoteHistoryRequest{
				Symbol:   "VCB",
				Start:    time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC),
				End:      time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
				Interval: "1m",
			},
			mockStatusCode: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data_1P": []map[string]interface{}{
					{
						"t": "2024-01-15 09:30:00",
						"o": 95000.0,
						"h": 95500.0,
						"l": 94800.0,
						"c": 95200.0,
						"v": 50000,
					},
				},
			},
			wantErr:        false,
			wantQuoteCount: 1,
			validateQuote: func(t *testing.T, quote vnstock.Quote) {
				if quote.Symbol != "VCB" {
					t.Errorf("Quote.Symbol = %v, want VCB", quote.Symbol)
				}
				if quote.Interval != "1m" {
					t.Errorf("Quote.Interval = %v, want 1m", quote.Interval)
				}
			},
		},
		{
			name: "empty symbol error",
			req: vnstock.QuoteHistoryRequest{
				Symbol:   "",
				Start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:      time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
				Interval: "1D",
			},
			wantErr: true,
			errCode: vnstock.InvalidInput,
		},
		{
			name: "invalid date range error",
			req: vnstock.QuoteHistoryRequest{
				Symbol:   "VNM",
				Start:    time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
				End:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Interval: "1D",
			},
			wantErr: true,
			errCode: vnstock.InvalidInput,
		},
		{
			name: "invalid interval error",
			req: vnstock.QuoteHistoryRequest{
				Symbol:   "VNM",
				Start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:      time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
				Interval: "invalid",
			},
			wantErr: true,
			errCode: vnstock.InvalidInput,
		},
		{
			name: "empty response no data error",
			req: vnstock.QuoteHistoryRequest{
				Symbol:   "VNM",
				Start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:      time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
				Interval: "1D",
			},
			mockStatusCode: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data_day": []map[string]interface{}{},
			},
			wantErr: true,
			errCode: vnstock.NoData,
		},
		{
			name: "HTTP 500 error",
			req: vnstock.QuoteHistoryRequest{
				Symbol:   "VNM",
				Start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:      time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
				Interval: "1D",
			},
			mockStatusCode: http.StatusInternalServerError,
			mockResponse: map[string]interface{}{
				"error": "internal server error",
			},
			wantErr: true,
			errCode: vnstock.HTTPError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip mock server setup for validation errors
			if tt.errCode == vnstock.InvalidInput {
				logger := slog.Default()
				c := New(http.DefaultClient, logger)
				ctx := context.Background()
				_, err := c.QuoteHistory(ctx, tt.req)

				if (err != nil) != tt.wantErr {
					t.Errorf("QuoteHistory() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if err != nil {
					var vErr *vnstock.Error
					if !vnstock.AsError(err, &vErr) {
						t.Errorf("QuoteHistory() error is not *vnstock.Error: %v", err)
						return
					}
					if vErr.Code != tt.errCode {
						t.Errorf("QuoteHistory() error code = %v, want %v", vErr.Code, tt.errCode)
					}
				}
				return
			}

			// Create mock server for API tests
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}

				// Verify URL contains symbol
				if !contains(r.URL.Path, tt.req.Symbol) {
					t.Errorf("Expected URL to contain symbol %s, got %s", tt.req.Symbol, r.URL.Path)
				}

				w.WriteHeader(tt.mockStatusCode)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Create connector with mock server
			logger := slog.Default()
			c := &Connector{
				client: server.Client(),
				logger: logger,
			}

			// Override the base URL for testing by modifying the request
			ctx := context.Background()

			// We need to intercept the URL construction, so we'll use a custom transport
			originalTransport := c.client.Transport
			c.client.Transport = &mockTransport{
				baseURL:  server.URL,
				original: originalTransport,
			}

			quotes, err := c.QuoteHistory(ctx, tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("QuoteHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				var vErr *vnstock.Error
				if !vnstock.AsError(err, &vErr) {
					t.Errorf("QuoteHistory() error is not *vnstock.Error: %v", err)
					return
				}
				if vErr.Code != tt.errCode {
					t.Errorf("QuoteHistory() error code = %v, want %v", vErr.Code, tt.errCode)
				}
				if tt.errCode == vnstock.HTTPError && vErr.StatusCode != tt.mockStatusCode {
					t.Errorf("QuoteHistory() status code = %v, want %v", vErr.StatusCode, tt.mockStatusCode)
				}
			} else {
				if len(quotes) != tt.wantQuoteCount {
					t.Errorf("QuoteHistory() returned %d quotes, want %d", len(quotes), tt.wantQuoteCount)
				}
				if tt.validateQuote != nil && len(quotes) > 0 {
					tt.validateQuote(t, quotes[0])
				}
			}
		})
	}
}

// mockTransport is a custom RoundTripper that replaces the base URL in requests.
type mockTransport struct {
	baseURL  string
	original http.RoundTripper
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Replace the KBS base URL with our mock server URL
	if contains(req.URL.String(), "kbbuddywts.kbsec.com.vn") {
		req.URL.Scheme = "http"
		req.URL.Host = req.URL.Host // Keep the original for path matching
		// Just use the mock server directly
		newURL := m.baseURL + req.URL.Path
		if req.URL.RawQuery != "" {
			newURL += "?" + req.URL.RawQuery
		}
		newReq, err := http.NewRequest(req.Method, newURL, req.Body)
		if err != nil {
			return nil, err
		}
		newReq.Header = req.Header
		if m.original != nil {
			return m.original.RoundTrip(newReq)
		}
		return http.DefaultTransport.RoundTrip(newReq)
	}
	if m.original != nil {
		return m.original.RoundTrip(req)
	}
	return http.DefaultTransport.RoundTrip(req)
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestMapIntervalToKBS tests the interval mapping function.
func TestMapIntervalToKBS(t *testing.T) {
	tests := []struct {
		name     string
		interval string
		want     string
		wantErr  bool
	}{
		{"1 minute", "1m", "1P", false},
		{"5 minutes", "5m", "5P", false},
		{"15 minutes", "15m", "15P", false},
		{"30 minutes", "30m", "30P", false},
		{"1 hour", "1H", "60P", false},
		{"1 day", "1D", "day", false},
		{"1 week", "1W", "week", false},
		{"1 month", "1M", "month", false},
		{"invalid interval", "invalid", "", true},
		{"empty interval", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapIntervalToKBS(tt.interval)
			if (err != nil) != tt.wantErr {
				t.Errorf("mapIntervalToKBS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("mapIntervalToKBS() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseKBSTimestamp tests the timestamp parsing function.
func TestParseKBSTimestamp(t *testing.T) {
	tests := []struct {
		name           string
		timeStr        string
		intervalSuffix string
		wantErr        bool
	}{
		{"daily format", "2024-01-15", "day", false},
		{"weekly format", "2024-01-15", "week", false},
		{"monthly format", "2024-01-15", "month", false},
		{"intraday format", "2024-01-15 09:30:00", "1P", false},
		{"intraday format 5P", "2024-01-15 14:27:23", "5P", false},
		{"invalid daily format", "15-01-2024", "day", true},
		{"invalid intraday format", "2024-01-15", "1P", true},
		{"empty time string", "", "day", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseKBSTimestamp(tt.timeStr, tt.intervalSuffix)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseKBSTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.IsZero() {
				t.Error("parseKBSTimestamp() returned zero time without error")
			}
		})
	}
}
