// Package vnd implements the Connector interface for VNDirect dchart API.
package vnd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	vnstock "github.com/dda10/vnstock-go"
)

func init() {
	vnstock.RegisterConnector("VND", func(client *http.Client, logger *slog.Logger) vnstock.Connector {
		return New(client, logger)
	})
}

// udfHistoryResponse represents the TradingView UDF history response format.
// Used internally to parse /dchart/history API responses.
type udfHistoryResponse struct {
	S string    `json:"s"` // status: "ok", "no_data", "error"
	T []int64   `json:"t"` // timestamps (Unix seconds)
	O []float64 `json:"o"` // open prices
	H []float64 `json:"h"` // high prices
	L []float64 `json:"l"` // low prices
	C []float64 `json:"c"` // close prices
	V []int64   `json:"v"` // volumes
}

// udfSymbolResponse represents the TradingView UDF symbol info response.
// Used internally to parse /dchart/symbols API responses.
type udfSymbolResponse struct {
	Symbol      string `json:"symbol"`
	Exchange    string `json:"exchange"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// udfSearchResult represents a single search result item.
// Used internally to parse /dchart/search API responses.
type udfSearchResult struct {
	Symbol      string `json:"symbol"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Exchange    string `json:"exchange"`
	Type        string `json:"type"`
}

const (
	baseURL      = "https://dchart-api.vndirect.com.vn"
	finfoBaseURL = "https://finfo-api.vndirect.com.vn"
)

// intervalToResolution maps vnstock interval strings to dchart resolution format.
var intervalToResolution = map[string]string{
	"1m":  "1",
	"5m":  "5",
	"15m": "15",
	"30m": "30",
	"1H":  "60",
	"1D":  "D",
	"1W":  "W",
	"1M":  "M",
}

// mapIntervalToResolution converts a vnstock interval string to dchart resolution format.
// Returns an error for unsupported intervals.
func mapIntervalToResolution(interval string) (string, error) {
	resolution, ok := intervalToResolution[interval]
	if !ok {
		return "", &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: fmt.Sprintf("unsupported interval: %s", interval),
		}
	}
	return resolution, nil
}

// Connector implements the vnstock.Connector interface for VNDirect dchart data source.
type Connector struct {
	client *http.Client
	logger *slog.Logger
}

// New creates a new VND connector with the provided HTTP client and logger.
func New(client *http.Client, logger *slog.Logger) *Connector {
	if logger == nil {
		logger = slog.Default()
	}
	return &Connector{
		client: client,
		logger: logger,
	}
}

// logRequest logs an HTTP request at DEBUG level with structured fields.
func (c *Connector) logRequest(method, url string, statusCode int, elapsed time.Duration) {
	c.logger.Debug("VND request",
		slog.String("method", method),
		slog.String("url", url),
		slog.Int("status", statusCode),
		slog.Duration("elapsed", elapsed),
	)
}

// doRequest performs an HTTP GET request to the dchart API and logs it.
// The endpoint parameter should be the path (e.g., "/dchart/history").
// Query parameters should be passed via the params argument.
func (c *Connector) doRequest(ctx context.Context, endpoint string, params url.Values) (*http.Response, error) {
	reqURL := baseURL + endpoint
	if len(params) > 0 {
		reqURL = reqURL + "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.NetworkError,
			Message: "failed to create request",
			Cause:   err,
		}
	}

	// Set headers for dchart API
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36")

	start := time.Now()
	resp, err := c.client.Do(req)
	elapsed := time.Since(start)

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}
	c.logRequest(http.MethodGet, reqURL, statusCode, elapsed)

	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.NetworkError,
			Message: "request failed",
			Cause:   err,
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &vnstock.Error{
			Code:       vnstock.HTTPError,
			Message:    fmt.Sprintf("HTTP error: %s", string(body)),
			StatusCode: resp.StatusCode,
		}
	}

	return resp, nil
}

// parseHistoryResponse parses a TradingView UDF history response into a slice of Quotes.
// It validates array lengths, handles status codes, and converts Unix timestamps to time.Time.
func parseHistoryResponse(body []byte, symbol, interval string) ([]vnstock.Quote, error) {
	var resp udfHistoryResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode response",
			Cause:   err,
		}
	}

	// Handle "no_data" status - return empty slice with no error
	if resp.S == "no_data" {
		return []vnstock.Quote{}, nil
	}

	// Handle "error" status - return NoData error
	if resp.S == "error" {
		return nil, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "API returned error status",
		}
	}

	// Validate array lengths are consistent
	n := len(resp.T)
	if len(resp.O) != n || len(resp.H) != n || len(resp.L) != n || len(resp.C) != n || len(resp.V) != n {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "response arrays have inconsistent lengths",
		}
	}

	// Convert to []vnstock.Quote
	quotes := make([]vnstock.Quote, n)
	for i := 0; i < n; i++ {
		quotes[i] = vnstock.Quote{
			Symbol:    symbol,
			Timestamp: time.Unix(resp.T[i], 0),
			Open:      resp.O[i],
			High:      resp.H[i],
			Low:       resp.L[i],
			Close:     resp.C[i],
			Volume:    resp.V[i],
			Interval:  interval,
		}
	}

	return quotes, nil
}

// QuoteHistory retrieves historical OHLCV data for a symbol from VNDirect dchart API.
func (c *Connector) QuoteHistory(ctx context.Context, req vnstock.QuoteHistoryRequest) ([]vnstock.Quote, error) {
	if req.Symbol == "" {
		return nil, &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: "symbol cannot be empty",
		}
	}
	if req.Start.After(req.End) {
		return nil, &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: "start date must be before or equal to end date",
		}
	}

	resolution, err := mapIntervalToResolution(req.Interval)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("symbol", req.Symbol)
	params.Set("resolution", resolution)
	params.Set("from", fmt.Sprintf("%d", req.Start.Unix()))
	params.Set("to", fmt.Sprintf("%d", req.End.Unix()))

	resp, err := c.doRequest(ctx, "/dchart/history", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.NetworkError,
			Message: "failed to read response body",
			Cause:   err,
		}
	}

	return parseHistoryResponse(body, req.Symbol, req.Interval)
}

// RealTimeQuotes returns ErrNotSupported as VNDirect dchart API does not support real-time quotes.
func (c *Connector) RealTimeQuotes(ctx context.Context, symbols []string) ([]vnstock.Quote, error) {
	return nil, vnstock.ErrNotSupported
}

// Listing retrieves the list of symbols by searching the dchart API.
// If exchange is empty, returns all symbols. Otherwise filters by exchange.
func (c *Connector) Listing(ctx context.Context, exchange string) ([]vnstock.ListingRecord, error) {
	params := url.Values{}
	params.Set("query", "")
	params.Set("limit", "30")
	if exchange != "" {
		params.Set("exchange", exchange)
	}

	resp, err := c.doRequest(ctx, "/dchart/search", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.NetworkError,
			Message: "failed to read response body",
			Cause:   err,
		}
	}

	var results []udfSearchResult
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode response",
			Cause:   err,
		}
	}

	if len(results) == 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "no listing data available",
		}
	}

	records := make([]vnstock.ListingRecord, len(results))
	for i, r := range results {
		records[i] = vnstock.ListingRecord{
			Symbol:      r.Symbol,
			Exchange:    r.Exchange,
			CompanyName: r.Description,
		}
	}

	return records, nil
}

// IndexCurrent returns ErrNotSupported as VNDirect dchart API does not support index data.
func (c *Connector) IndexCurrent(ctx context.Context, name string) (vnstock.IndexRecord, error) {
	return vnstock.IndexRecord{}, vnstock.ErrNotSupported
}

// IndexHistory returns ErrNotSupported as VNDirect dchart API does not support index data.
func (c *Connector) IndexHistory(ctx context.Context, req vnstock.IndexHistoryRequest) ([]vnstock.IndexRecord, error) {
	return nil, vnstock.ErrNotSupported
}

// CompanyProfile returns ErrNotSupported as VNDirect dchart API does not support company profiles.
func (c *Connector) CompanyProfile(ctx context.Context, symbol string) (vnstock.CompanyProfile, error) {
	return vnstock.CompanyProfile{}, vnstock.ErrNotSupported
}

// Officers returns ErrNotSupported as VNDirect dchart API does not support officer data.
func (c *Connector) Officers(ctx context.Context, symbol string) ([]vnstock.Officer, error) {
	return nil, vnstock.ErrNotSupported
}

// FinancialStatement returns ErrNotSupported as VNDirect dchart API does not support financial statements.
func (c *Connector) FinancialStatement(ctx context.Context, req vnstock.FinancialRequest) ([]vnstock.FinancialPeriod, error) {
	return nil, vnstock.ErrNotSupported
}

// FinancialRatios retrieves key financial ratios from VNDirect finfo API.
// Uses the same API endpoints discovered in vnquant.
func (c *Connector) FinancialRatios(ctx context.Context, req vnstock.FinancialRatioRequest) (vnstock.FinancialRatio, error) {
	if req.Symbol == "" {
		return vnstock.FinancialRatio{}, &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: "symbol cannot be empty",
		}
	}

	// Default to latest report date if not specified
	reportDate := req.ReportDate
	if reportDate.IsZero() {
		// Use end of previous year as default
		now := time.Now()
		reportDate = time.Date(now.Year()-1, 12, 31, 0, 0, 0, 0, time.UTC)
	}

	// Item codes from vnquant for basic financial ratios:
	// 53030 - ROA, 52005 - ROE, 51050 - Net Profit Margin
	// 53021 - Revenue Growth, 52001/52002 - Profit Growth
	// 54018 - EPS, 712010-712040 - Various ratios
	itemCodes := "53030,52005,51050,53021,52001,52002,54018,712010,712020,712030,712040"

	// Build query string
	query := fmt.Sprintf("code:%s~itemCode:%s~reportDate:%s",
		req.Symbol,
		itemCodes,
		reportDate.Format("2006-01-02"),
	)

	params := url.Values{}
	params.Set("q", query)

	reqURL := finfoBaseURL + "/v4/ratios?" + params.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return vnstock.FinancialRatio{}, &vnstock.Error{
			Code:    vnstock.NetworkError,
			Message: "failed to create request",
			Cause:   err,
		}
	}

	// Set headers
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36")

	start := time.Now()
	resp, err := c.client.Do(httpReq)
	elapsed := time.Since(start)

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}
	c.logRequest(http.MethodGet, reqURL, statusCode, elapsed)

	if err != nil {
		return vnstock.FinancialRatio{}, &vnstock.Error{
			Code:    vnstock.NetworkError,
			Message: "request failed",
			Cause:   err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return vnstock.FinancialRatio{}, &vnstock.Error{
			Code:       vnstock.HTTPError,
			Message:    fmt.Sprintf("HTTP error: %s", string(body)),
			StatusCode: resp.StatusCode,
		}
	}

	// Parse response
	var apiResp struct {
		Data []struct {
			Code       string  `json:"code"`
			ItemCode   string  `json:"itemCode"`
			ItemName   string  `json:"itemName"`
			Value      float64 `json:"value"`
			ReportDate string  `json:"reportDate"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return vnstock.FinancialRatio{}, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode response",
			Cause:   err,
		}
	}

	if len(apiResp.Data) == 0 {
		return vnstock.FinancialRatio{}, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "no financial ratio data available",
		}
	}

	// Map item codes to ratio fields
	ratio := vnstock.FinancialRatio{
		Symbol:     req.Symbol,
		ReportDate: reportDate,
	}

	for _, item := range apiResp.Data {
		switch item.ItemCode {
		case "53030":
			ratio.ROA = item.Value
		case "52005":
			ratio.ROE = item.Value
		case "51050":
			ratio.NetProfitMargin = item.Value
		case "53021":
			ratio.RevenueGrowth = item.Value
		case "52001", "52002":
			ratio.ProfitGrowth = item.Value
		case "54018":
			ratio.EPS = item.Value
		case "712010":
			ratio.PE = item.Value
		case "712020":
			ratio.PB = item.Value
		case "712030":
			ratio.CurrentRatio = item.Value
		case "712040":
			ratio.DebtToEquity = item.Value
		}
	}

	return ratio, nil
}
