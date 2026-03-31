// Package vci implements the Connector interface for the VCI data source.
package vci

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	vnstock "github.com/dda10/vnstock-go"
)

func init() {
	vnstock.RegisterConnector("VCI", func(client *http.Client, logger *slog.Logger) vnstock.Connector {
		return New(client, logger)
	})
}

const (
	baseURL = "https://trading.vietcap.com.vn/api"
)

// Connector implements the vnstock.Connector interface for VCI data source.
type Connector struct {
	client *http.Client
	logger *slog.Logger
}

// New creates a new VCI connector with the provided HTTP client and logger.
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
	c.logger.Debug("VCI request",
		slog.String("method", method),
		slog.String("url", url),
		slog.Int("status", statusCode),
		slog.Duration("elapsed", elapsed),
	)
}

// doRequest performs an HTTP request with JSON payload and logs it.
func (c *Connector) doRequest(ctx context.Context, method, endpoint string, payload interface{}) (*http.Response, error) {
	reqURL := baseURL + endpoint

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, &vnstock.Error{
				Code:    vnstock.SerialiseError,
				Message: "failed to marshal request payload",
				Cause:   err,
			}
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.NetworkError,
			Message: "failed to create request",
			Cause:   err,
		}
	}

	// Set headers to bypass anti-bot protection
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,vi-VN;q=0.8,vi;q=0.7")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("DNT", "1")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://trading.vietcap.com.vn/")
	req.Header.Set("Origin", "https://trading.vietcap.com.vn/")

	start := time.Now()
	resp, err := c.client.Do(req)
	elapsed := time.Since(start)

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}
	c.logRequest(method, reqURL, statusCode, elapsed)

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

// mapIntervalToTimeFrame converts vnstock interval to VCI timeFrame.
func mapIntervalToTimeFrame(interval string) string {
	switch interval {
	case "1m", "5m", "15m", "30m":
		return "ONE_MINUTE"
	case "1H":
		return "ONE_HOUR"
	case "1D", "1W", "1M":
		return "ONE_DAY"
	default:
		return "ONE_DAY"
	}
}

// QuoteHistory retrieves historical OHLCV data for a symbol.
// QuoteHistory retrieves historical OHLCV data for a symbol.
func (c *Connector) QuoteHistory(ctx context.Context, req vnstock.QuoteHistoryRequest) ([]vnstock.Quote, error) {
	// Validate request
	if req.Start.After(req.End) {
		return nil, &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: "start date must be before or equal to end date",
		}
	}

	// Calculate countBack (number of periods)
	countBack := int(req.End.Sub(req.Start).Hours() / 24)
	if countBack <= 0 {
		countBack = 100
	}

	// Prepare payload for VCI API
	payload := map[string]interface{}{
		"timeFrame": mapIntervalToTimeFrame(req.Interval),
		"symbols":   []string{req.Symbol},
		"to":        req.End.Unix(),
		"countBack": countBack,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/chart/OHLCChart/gap-chart", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read raw response to handle flexible JSON format
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to read response body",
			Cause:   err,
		}
	}

	// VCI returns data in array format with mixed types
	var result []map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode response",
			Cause:   err,
		}
	}

	if len(result) == 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "no data returned from API",
		}
	}

	// Extract arrays from the first element
	symbolData := result[0]

	// Helper function to convert interface{} arrays to typed arrays
	toInt64Array := func(v interface{}) ([]int64, error) {
		arr, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected array")
		}
		result := make([]int64, len(arr))
		for i, item := range arr {
			switch val := item.(type) {
			case float64:
				result[i] = int64(val)
			case string:
				// Try parsing string as int64
				parsed, err := fmt.Sscanf(val, "%d", &result[i])
				if err != nil || parsed != 1 {
					return nil, fmt.Errorf("failed to parse timestamp: %v", val)
				}
			default:
				return nil, fmt.Errorf("unexpected type for timestamp: %T", item)
			}
		}
		return result, nil
	}

	toFloat64Array := func(v interface{}) ([]float64, error) {
		arr, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected array")
		}
		result := make([]float64, len(arr))
		for i, item := range arr {
			switch val := item.(type) {
			case float64:
				result[i] = val
			case string:
				// Try parsing string as float64
				parsed, err := fmt.Sscanf(val, "%f", &result[i])
				if err != nil || parsed != 1 {
					return nil, fmt.Errorf("failed to parse float: %v", val)
				}
			default:
				return nil, fmt.Errorf("unexpected type for float: %T", item)
			}
		}
		return result, nil
	}

	// Parse each field
	timestamps, err := toInt64Array(symbolData["t"])
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse timestamps",
			Cause:   err,
		}
	}

	opens, err := toFloat64Array(symbolData["o"])
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse open prices",
			Cause:   err,
		}
	}

	highs, err := toFloat64Array(symbolData["h"])
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse high prices",
			Cause:   err,
		}
	}

	lows, err := toFloat64Array(symbolData["l"])
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse low prices",
			Cause:   err,
		}
	}

	closes, err := toFloat64Array(symbolData["c"])
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse close prices",
			Cause:   err,
		}
	}

	volumes, err := toInt64Array(symbolData["v"])
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse volumes",
			Cause:   err,
		}
	}

	// Transform to quotes
	numPoints := len(timestamps)
	quotes := make([]vnstock.Quote, 0, numPoints)
	for i := 0; i < numPoints; i++ {
		quotes = append(quotes, vnstock.Quote{
			Symbol:    req.Symbol,
			Timestamp: time.Unix(timestamps[i], 0),
			Open:      opens[i],
			High:      highs[i],
			Low:       lows[i],
			Close:     closes[i],
			Volume:    volumes[i],
			Interval:  req.Interval,
		})
	}

	return quotes, nil
}

// RealTimeQuotes retrieves the most recent quote for one or more symbols.
func (c *Connector) RealTimeQuotes(ctx context.Context, symbols []string) ([]vnstock.Quote, error) {
	if len(symbols) == 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: "symbols list cannot be empty",
		}
	}

	// Use intraday endpoint for real-time data
	quotes := make([]vnstock.Quote, 0, len(symbols))

	for _, symbol := range symbols {
		payload := map[string]interface{}{
			"symbol":    symbol,
			"limit":     1,
			"truncTime": time.Now().Unix(),
		}

		resp, err := c.doRequest(ctx, http.MethodPost, "/market-watch/LEData/getAll", payload)
		if err != nil {
			c.logger.Warn("failed to get real-time quote", slog.String("symbol", symbol), slog.Any("error", err))
			continue
		}

		var result []struct {
			T int64   `json:"t"` // timestamp
			O float64 `json:"o"` // open
			H float64 `json:"h"` // high
			L float64 `json:"l"` // low
			C float64 `json:"c"` // close
			V int64   `json:"v"` // volume
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			c.logger.Warn("failed to decode real-time quote", slog.String("symbol", symbol), slog.Any("error", err))
			continue
		}
		resp.Body.Close()

		if len(result) > 0 {
			item := result[0]
			quotes = append(quotes, vnstock.Quote{
				Symbol:    symbol,
				Timestamp: time.Unix(item.T, 0),
				Open:      item.O,
				High:      item.H,
				Low:       item.L,
				Close:     item.C,
				Volume:    item.V,
				Interval:  "realtime",
			})
		}
	}

	if len(quotes) == 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "no real-time data available for any symbol",
		}
	}

	return quotes, nil
}

// vciListingItem represents a single item from the VCI listing endpoint.
// The VCI API returns camelCase JSON fields.
type vciListingItem struct {
	Symbol         string `json:"symbol"`
	Board          string `json:"board"`          // exchange (HOSE, HNX, UPCOM)
	Type           string `json:"type"`           // STOCK, ETF, CW, BOND, etc.
	OrganName      string `json:"organName"`      // Vietnamese company name
	EnOrganName    string `json:"enOrganName"`    // English company name
	OrganShortName string `json:"organShortName"` // Short name
	IcbName3       string `json:"icbName3"`       // ICB industry level 3 (Vietnamese)
	EnIcbName3     string `json:"enIcbName3"`     // ICB industry level 3 (English)
	IcbName4       string `json:"icbName4"`       // ICB industry level 4 (Vietnamese)
	EnIcbName4     string `json:"enIcbName4"`     // ICB industry level 4 (English)
}

// Listing retrieves the full list of symbols traded on an exchange.
// It uses the VCI REST endpoint GET /price/symbols/getAll.
// If exchange is empty, returns all symbols. If specified, filters by exchange.
func (c *Connector) Listing(ctx context.Context, exchange string) ([]vnstock.ListingRecord, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/price/symbols/getAll", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var items []vciListingItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode listing response",
			Cause:   err,
		}
	}

	records := make([]vnstock.ListingRecord, 0, len(items))
	for _, item := range items {
		// Filter by exchange if specified
		if exchange != "" && !strings.EqualFold(item.Board, exchange) {
			continue
		}

		// Prefer English company name, fall back to Vietnamese
		companyName := item.EnOrganName
		if companyName == "" {
			companyName = item.OrganName
		}

		// Use ICB industry name as sector (prefer English)
		sector := item.EnIcbName3
		if sector == "" {
			sector = item.IcbName3
		}

		records = append(records, vnstock.ListingRecord{
			Symbol:      item.Symbol,
			Exchange:    item.Board,
			CompanyName: companyName,
			Sector:      sector,
		})
	}

	if len(records) == 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "no listing data available",
		}
	}

	return records, nil
}

// vciIndexNameMap maps common index names to VCI API index codes.
var vciIndexNameMap = map[string]string{
	"VN-Index":    "VNINDEX",
	"HNX-Index":   "HNXINDEX",
	"UPCOM-Index": "UPCOMINDEX",
}

// mapVCIIndexName converts a user-facing index name to the VCI API code.
// Returns the mapped name and true if valid, or empty string and false if unrecognized.
func mapVCIIndexName(name string) (string, bool) {
	mapped, ok := vciIndexNameMap[name]
	return mapped, ok
}

// doGraphQLRequest performs a GraphQL query to the VCI GraphQL endpoint.
func (c *Connector) doGraphQLRequest(ctx context.Context, query string, variables map[string]interface{}) (*http.Response, error) {
	graphqlURL := "https://trading.vietcap.com.vn/data-mt/graphql"

	payload := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to marshal GraphQL request",
			Cause:   err,
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, graphqlURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.NetworkError,
			Message: "failed to create GraphQL request",
			Cause:   err,
		}
	}

	// Set headers similar to doRequest
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,vi-VN;q=0.8,vi;q=0.7")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("DNT", "1")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://trading.vietcap.com.vn/")
	req.Header.Set("Origin", "https://trading.vietcap.com.vn/")

	start := time.Now()
	resp, err := c.client.Do(req)
	elapsed := time.Since(start)

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}
	c.logRequest(http.MethodPost, graphqlURL, statusCode, elapsed)

	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.NetworkError,
			Message: "GraphQL request failed",
			Cause:   err,
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &vnstock.Error{
			Code:       vnstock.HTTPError,
			Message:    fmt.Sprintf("GraphQL HTTP error: %s", string(body)),
			StatusCode: resp.StatusCode,
		}
	}

	return resp, nil
}

// IndexCurrent retrieves the current value of a named market index.
func (c *Connector) IndexCurrent(ctx context.Context, name string) (vnstock.IndexRecord, error) {
	vciName, ok := mapVCIIndexName(name)
	if !ok {
		return vnstock.IndexRecord{}, &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: fmt.Sprintf("unrecognized index name: %s", name),
		}
	}

	// Use the same OHLC endpoint as quotes but for index
	today := time.Now()
	payload := map[string]interface{}{
		"timeFrame": "ONE_DAY",
		"symbols":   []string{vciName},
		"to":        today.Unix(),
		"countBack": 1,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/chart/OHLCChart/gap-chart", payload)
	if err != nil {
		return vnstock.IndexRecord{}, err
	}
	defer resp.Body.Close()

	// Read raw response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return vnstock.IndexRecord{}, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to read index response body",
			Cause:   err,
		}
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return vnstock.IndexRecord{}, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode index response",
			Cause:   err,
		}
	}

	if len(result) == 0 {
		return vnstock.IndexRecord{}, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "no index data returned from API",
		}
	}

	indexData := result[0]

	// Helper functions to parse arrays
	toInt64Array := func(v interface{}) ([]int64, error) {
		arr, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected array")
		}
		result := make([]int64, len(arr))
		for i, item := range arr {
			switch val := item.(type) {
			case float64:
				result[i] = int64(val)
			default:
				return nil, fmt.Errorf("unexpected type for timestamp: %T", item)
			}
		}
		return result, nil
	}

	toFloat64Array := func(v interface{}) ([]float64, error) {
		arr, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected array")
		}
		result := make([]float64, len(arr))
		for i, item := range arr {
			switch val := item.(type) {
			case float64:
				result[i] = val
			default:
				return nil, fmt.Errorf("unexpected type for float: %T", item)
			}
		}
		return result, nil
	}

	timestamps, err := toInt64Array(indexData["t"])
	if err != nil || len(timestamps) == 0 {
		return vnstock.IndexRecord{}, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse index timestamps",
			Cause:   err,
		}
	}

	opens, err := toFloat64Array(indexData["o"])
	if err != nil || len(opens) == 0 {
		return vnstock.IndexRecord{}, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse index open prices",
			Cause:   err,
		}
	}

	highs, err := toFloat64Array(indexData["h"])
	if err != nil || len(highs) == 0 {
		return vnstock.IndexRecord{}, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse index high prices",
			Cause:   err,
		}
	}

	lows, err := toFloat64Array(indexData["l"])
	if err != nil || len(lows) == 0 {
		return vnstock.IndexRecord{}, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse index low prices",
			Cause:   err,
		}
	}

	closes, err := toFloat64Array(indexData["c"])
	if err != nil || len(closes) == 0 {
		return vnstock.IndexRecord{}, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse index close prices",
			Cause:   err,
		}
	}

	volumes, err := toInt64Array(indexData["v"])
	if err != nil || len(volumes) == 0 {
		return vnstock.IndexRecord{}, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse index volumes",
			Cause:   err,
		}
	}

	// Get the latest (last) data point
	idx := len(timestamps) - 1

	return vnstock.IndexRecord{
		Name:      name,
		Timestamp: time.Unix(timestamps[idx], 0),
		Value:     closes[idx],
		Open:      opens[idx],
		High:      highs[idx],
		Low:       lows[idx],
		Close:     closes[idx],
		Volume:    volumes[idx],
	}, nil
}

// IndexHistory retrieves historical values for a named market index.
func (c *Connector) IndexHistory(ctx context.Context, req vnstock.IndexHistoryRequest) ([]vnstock.IndexRecord, error) {
	vciName, ok := mapVCIIndexName(req.Name)
	if !ok {
		return nil, &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: fmt.Sprintf("unrecognized index name: %s", req.Name),
		}
	}

	// Validate date range
	if req.Start.After(req.End) {
		return nil, &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: "start date must be before or equal to end date",
		}
	}

	// Calculate countBack (number of periods)
	countBack := int(req.End.Sub(req.Start).Hours() / 24)
	if countBack <= 0 {
		countBack = 100
	}

	// Use the same OHLC endpoint as quotes but for index
	payload := map[string]interface{}{
		"timeFrame": mapIntervalToTimeFrame(req.Interval),
		"symbols":   []string{vciName},
		"to":        req.End.Unix(),
		"countBack": countBack,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/chart/OHLCChart/gap-chart", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read raw response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to read index history response body",
			Cause:   err,
		}
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode index history response",
			Cause:   err,
		}
	}

	if len(result) == 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "no index history data returned from API",
		}
	}

	indexData := result[0]

	// Helper functions to parse arrays
	toInt64Array := func(v interface{}) ([]int64, error) {
		arr, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected array")
		}
		result := make([]int64, len(arr))
		for i, item := range arr {
			switch val := item.(type) {
			case float64:
				result[i] = int64(val)
			default:
				return nil, fmt.Errorf("unexpected type for timestamp: %T", item)
			}
		}
		return result, nil
	}

	toFloat64Array := func(v interface{}) ([]float64, error) {
		arr, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected array")
		}
		result := make([]float64, len(arr))
		for i, item := range arr {
			switch val := item.(type) {
			case float64:
				result[i] = val
			default:
				return nil, fmt.Errorf("unexpected type for float: %T", item)
			}
		}
		return result, nil
	}

	timestamps, err := toInt64Array(indexData["t"])
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse index timestamps",
			Cause:   err,
		}
	}

	opens, err := toFloat64Array(indexData["o"])
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse index open prices",
			Cause:   err,
		}
	}

	highs, err := toFloat64Array(indexData["h"])
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse index high prices",
			Cause:   err,
		}
	}

	lows, err := toFloat64Array(indexData["l"])
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse index low prices",
			Cause:   err,
		}
	}

	closes, err := toFloat64Array(indexData["c"])
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse index close prices",
			Cause:   err,
		}
	}

	volumes, err := toInt64Array(indexData["v"])
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse index volumes",
			Cause:   err,
		}
	}

	// Transform to IndexRecord slice
	numPoints := len(timestamps)
	records := make([]vnstock.IndexRecord, 0, numPoints)
	for i := 0; i < numPoints; i++ {
		records = append(records, vnstock.IndexRecord{
			Name:      req.Name,
			Timestamp: time.Unix(timestamps[i], 0),
			Value:     closes[i],
			Open:      opens[i],
			High:      highs[i],
			Low:       lows[i],
			Close:     closes[i],
			Volume:    volumes[i],
		})
	}

	return records, nil
}

// CompanyProfile retrieves descriptive information about a listed company.
func (c *Connector) CompanyProfile(ctx context.Context, symbol string) (vnstock.CompanyProfile, error) {
	if symbol == "" {
		return vnstock.CompanyProfile{}, &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: "symbol cannot be empty",
		}
	}

	// GraphQL query for company profile data
	query := `
		query CompanyProfile($symbol: String!) {
			companyProfile(symbol: $symbol) {
				symbol
				companyName
				exchange
				sector
				industry
				foundedYear
				website
				companyDescription
				address
				phone
				email
				charterCapital
				listedDate
				faceValue
				listedPrice
				listedVolume
				marketCap
				chairmanName
				chairmanPosition
			}
			shareholders(symbol: $symbol) {
				name
				shares
				percentage
				note
			}
			ownership(symbol: $symbol) {
				name
				shares
				percentage
			}
		}
	`

	variables := map[string]interface{}{
		"symbol": symbol,
	}

	resp, err := c.doGraphQLRequest(ctx, query, variables)
	if err != nil {
		return vnstock.CompanyProfile{}, err
	}
	defer resp.Body.Close()

	// Parse GraphQL response
	var graphqlResp struct {
		Data struct {
			CompanyProfile struct {
				Symbol             string  `json:"symbol"`
				CompanyName        string  `json:"companyName"`
				Exchange           string  `json:"exchange"`
				Sector             string  `json:"sector"`
				Industry           string  `json:"industry"`
				FoundedYear        string  `json:"foundedYear"`
				Website            string  `json:"website"`
				CompanyDescription string  `json:"companyDescription"`
				Address            string  `json:"address"`
				Phone              string  `json:"phone"`
				Email              string  `json:"email"`
				CharterCapital     float64 `json:"charterCapital"`
				ListedDate         string  `json:"listedDate"`
				FaceValue          float64 `json:"faceValue"`
				ListedPrice        float64 `json:"listedPrice"`
				ListedVolume       int64   `json:"listedVolume"`
				MarketCap          float64 `json:"marketCap"`
				ChairmanName       string  `json:"chairmanName"`
				ChairmanPosition   string  `json:"chairmanPosition"`
			} `json:"companyProfile"`
			Shareholders []struct {
				Name       string  `json:"name"`
				Shares     float64 `json:"shares"`
				Percentage float64 `json:"percentage"`
				Note       string  `json:"note"`
			} `json:"shareholders"`
			Ownership []struct {
				Name       string  `json:"name"`
				Shares     float64 `json:"shares"`
				Percentage float64 `json:"percentage"`
			} `json:"ownership"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&graphqlResp); err != nil {
		return vnstock.CompanyProfile{}, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode company profile response",
			Cause:   err,
		}
	}

	// Check for GraphQL errors
	if len(graphqlResp.Errors) > 0 {
		return vnstock.CompanyProfile{}, &vnstock.Error{
			Code:    vnstock.HTTPError,
			Message: fmt.Sprintf("GraphQL error: %s", graphqlResp.Errors[0].Message),
		}
	}

	// Map to CompanyProfile struct
	profile := vnstock.CompanyProfile{
		Symbol:           graphqlResp.Data.CompanyProfile.Symbol,
		Name:             graphqlResp.Data.CompanyProfile.CompanyName,
		Exchange:         graphqlResp.Data.CompanyProfile.Exchange,
		Sector:           graphqlResp.Data.CompanyProfile.Sector,
		Industry:         graphqlResp.Data.CompanyProfile.Industry,
		Founded:          graphqlResp.Data.CompanyProfile.FoundedYear,
		Website:          graphqlResp.Data.CompanyProfile.Website,
		Description:      graphqlResp.Data.CompanyProfile.CompanyDescription,
		Address:          graphqlResp.Data.CompanyProfile.Address,
		Phone:            graphqlResp.Data.CompanyProfile.Phone,
		Email:            graphqlResp.Data.CompanyProfile.Email,
		CharterCapital:   graphqlResp.Data.CompanyProfile.CharterCapital,
		ListedDate:       graphqlResp.Data.CompanyProfile.ListedDate,
		FaceValue:        graphqlResp.Data.CompanyProfile.FaceValue,
		ListedPrice:      graphqlResp.Data.CompanyProfile.ListedPrice,
		ListedVolume:     graphqlResp.Data.CompanyProfile.ListedVolume,
		MarketCap:        graphqlResp.Data.CompanyProfile.MarketCap,
		ChairmanName:     graphqlResp.Data.CompanyProfile.ChairmanName,
		ChairmanPosition: graphqlResp.Data.CompanyProfile.ChairmanPosition,
	}

	// Map shareholders
	if len(graphqlResp.Data.Shareholders) > 0 {
		profile.Shareholders = make([]vnstock.Shareholder, len(graphqlResp.Data.Shareholders))
		for i, sh := range graphqlResp.Data.Shareholders {
			profile.Shareholders[i] = vnstock.Shareholder{
				Name:       sh.Name,
				Shares:     sh.Shares,
				Percentage: sh.Percentage,
				Note:       sh.Note,
			}
		}
	}

	// Map ownership
	if len(graphqlResp.Data.Ownership) > 0 {
		profile.Ownership = make([]vnstock.OwnershipEntry, len(graphqlResp.Data.Ownership))
		for i, own := range graphqlResp.Data.Ownership {
			profile.Ownership[i] = vnstock.OwnershipEntry{
				Name:       own.Name,
				Shares:     own.Shares,
				Percentage: own.Percentage,
			}
		}
	}

	// Validate required fields
	if profile.Symbol == "" || profile.Name == "" || profile.Exchange == "" || profile.Sector == "" {
		return vnstock.CompanyProfile{}, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "company profile missing required fields",
		}
	}

	return profile, nil
}

// Officers retrieves the list of officers and executives for a company.
func (c *Connector) Officers(ctx context.Context, symbol string) ([]vnstock.Officer, error) {
	if symbol == "" {
		return nil, &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: "symbol cannot be empty",
		}
	}

	// GraphQL query for officers data
	query := `
		query Officers($symbol: String!) {
			officers(symbol: $symbol) {
				name
				title
				appointmentDate
			}
		}
	`

	variables := map[string]interface{}{
		"symbol": symbol,
	}

	resp, err := c.doGraphQLRequest(ctx, query, variables)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse GraphQL response
	var graphqlResp struct {
		Data struct {
			Officers []struct {
				Name            string `json:"name"`
				Title           string `json:"title"`
				AppointmentDate string `json:"appointmentDate"`
			} `json:"officers"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&graphqlResp); err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode officers response",
			Cause:   err,
		}
	}

	// Check for GraphQL errors
	if len(graphqlResp.Errors) > 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.HTTPError,
			Message: fmt.Sprintf("GraphQL error: %s", graphqlResp.Errors[0].Message),
		}
	}

	// Map to Officer slice
	officers := make([]vnstock.Officer, 0, len(graphqlResp.Data.Officers))
	for _, off := range graphqlResp.Data.Officers {
		// Validate required fields
		if off.Name == "" || off.Title == "" {
			continue
		}
		officers = append(officers, vnstock.Officer{
			Name:            off.Name,
			Title:           off.Title,
			AppointmentDate: off.AppointmentDate,
		})
	}

	if len(officers) == 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "no officers data available",
		}
	}

	return officers, nil
}

// FinancialStatement retrieves financial statement data for a company.
// FinancialStatement retrieves financial statement data for a company.
// Uses the CompanyFinancialRatio GraphQL query (same as Python vnstock).
// The VCI API returns all financial data in a single query with field codes
// like BSA1 (balance sheet), ISA1 (income statement), CFA1 (cash flow).
func (c *Connector) FinancialStatement(ctx context.Context, req vnstock.FinancialRequest) ([]vnstock.FinancialPeriod, error) {
	if req.Symbol == "" {
		return nil, &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: "symbol cannot be empty",
		}
	}

	// Map period format: "annual" -> "Y", "quarter" -> "Q"
	period := req.Period
	switch period {
	case "annual", "yearly", "year":
		period = "Y"
	case "quarter", "quarterly":
		period = "Q"
	}

	// GraphQL query matching the Python vnstock implementation.
	// Uses CompanyFinancialRatio which returns all financial data including
	// income statement (ISA/ISB), balance sheet (BSA/BSB), cash flow (CFA/CFB),
	// and computed ratios (pe, pb, roe, roa, etc.)
	query := `
		fragment Ratios on CompanyFinancialRatio {
			ticker
			yearReport
			lengthReport
			revenue
			revenueGrowth
			netProfit
			netProfitGrowth
			ebitMargin
			roe
			roic
			roa
			pe
			pb
			eps
			currentRatio
			cashRatio
			quickRatio
			interestCoverage
			ae
			netProfitMargin
			grossMargin
			ev
			issueShare
			ps
			pcf
			bvps
			evPerEbitda
			BSA1
			BSA2
			BSA5
			BSA8
			BSA10
			BSA16
			BSA22
			BSA23
			BSA24
			BSA27
			BSA29
			BSA43
			BSA46
			BSA50
			BSA53
			BSA54
			BSA55
			BSA56
			BSA58
			BSA67
			BSA71
			BSA78
			BSA79
			BSA80
			BSA86
			BSA90
			BSA96
			CFA21
			CFA22
			at
			fat
			acp
			dso
			dpo
			ccc
			de
			le
			ebitda
			ebit
			dividend
			RTQ10
			charterCapitalRatio
			RTQ4
			epsTTM
			charterCapital
			fae
			RTQ17
			CFA26
			CFA6
			CFA9
			BSA85
			CFA36
			BSB98
			BSB101
			BSA89
			CFA34
			CFA14
			ISB34
			ISB27
			ISA23
			ISA102
			CFA27
			CFA12
			CFA28
			BSA18
			BSB102
			BSB110
			BSB108
			CFA23
			ISB41
			BSB103
			BSA40
			BSB99
			CFA16
			CFA18
			CFA3
			ISB30
			BSA33
			ISB29
			ISA2
			CFA24
			BSB105
			CFA37
			BSA95
			CFA10
			ISA4
			BSA82
			CFA25
			BSB111
			ISA20
			CFA19
			ISA6
			ISA3
			BSB100
			ISB31
			ISB38
			ISB26
			CFA20
			CFA35
			ISA17
			ISA9
			CFA4
			ISA7
			CFA5
			ISA22
			CFA8
			CFA33
			CFA29
			BSA30
			BSA84
			BSA44
			BSB107
			ISB37
			ISA8
			BSB109
			ISA19
			ISB36
			ISA13
			ISA1
			ISA14
			BSB112
			ISA21
			ISA10
			CFA11
			ISA12
			BSA15
			BSB104
			BSA92
			BSB106
			BSA94
			ISA18
			CFA17
			BSB114
			ISA15
			BSB116
			ISB28
			BSB97
			CFA15
			ISA11
			ISB33
			BSA47
			ISB40
			ISB39
			CFA7
			CFA13
			ISB25
			BSA45
			BSB118
			CFA1
			ISB35
			CFA31
			BSB113
			ISB32
			ISA16
			BSA48
			BSA36
			CFA30
			CFA2
			CFA38
			CFA32
			ISA5
			BSA49
			__typename
		}

		query Query($ticker: String!, $period: String!) {
			CompanyFinancialRatio(ticker: $ticker, period: $period) {
				ratio {
					...Ratios
					__typename
				}
				period
				__typename
			}
		}
	`

	variables := map[string]interface{}{
		"ticker": req.Symbol,
		"period": period,
	}

	resp, err := c.doGraphQLRequest(ctx, query, variables)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse the CompanyFinancialRatio response
	var graphqlResp struct {
		Data struct {
			CompanyFinancialRatio struct {
				Ratio []map[string]interface{} `json:"ratio"`
			} `json:"CompanyFinancialRatio"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&graphqlResp); err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode financial statement response",
			Cause:   err,
		}
	}

	if len(graphqlResp.Errors) > 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.HTTPError,
			Message: fmt.Sprintf("GraphQL error: %s", graphqlResp.Errors[0].Message),
		}
	}

	ratioData := graphqlResp.Data.CompanyFinancialRatio.Ratio
	if len(ratioData) == 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "no financial statement data available",
		}
	}

	// Convert ratio data to FinancialPeriod slice
	// Each ratio entry contains yearReport, lengthReport, and all financial fields
	periods := make([]vnstock.FinancialPeriod, 0, len(ratioData))
	for _, entry := range ratioData {
		year := int(toFloat64(entry["yearReport"]))
		quarter := int(toFloat64(entry["lengthReport"]))

		fields := make(map[string]float64)
		for key, val := range entry {
			switch key {
			case "ticker", "yearReport", "lengthReport", "updateDate", "__typename":
				continue
			default:
				if v := toFloat64(val); v != 0 {
					fields[key] = v
				}
			}
		}

		// Map common aliases for downstream consumers
		if _, ok := fields["pe"]; !ok {
			if v, ok := fields["pe_ratio"]; ok {
				fields["pe"] = v
			}
		}
		if _, ok := fields["market_cap"]; !ok {
			// market_cap = pe * netProfit (approximate)
			if pe, ok1 := fields["pe"]; ok1 {
				if np, ok2 := fields["netProfit"]; ok2 && np != 0 {
					fields["market_cap"] = pe * np
				}
			}
		}
		if _, ok := fields["ev_ebitda"]; !ok {
			if v, ok := fields["evPerEbitda"]; ok {
				fields["ev_ebitda"] = v
			}
		}
		if _, ok := fields["debt_to_equity"]; !ok {
			if v, ok := fields["de"]; ok {
				fields["debt_to_equity"] = v
			}
		}
		if _, ok := fields["dividend_yield"]; !ok {
			if v, ok := fields["dividend"]; ok {
				fields["dividend_yield"] = v
			}
		}
		if _, ok := fields["revenue_growth"]; !ok {
			if v, ok := fields["revenueGrowth"]; ok {
				fields["revenue_growth"] = v
			}
		}
		if _, ok := fields["profit_growth"]; !ok {
			if v, ok := fields["netProfitGrowth"]; ok {
				fields["profit_growth"] = v
			}
		}
		if _, ok := fields["net_income_growth"]; !ok {
			if v, ok := fields["netProfitGrowth"]; ok {
				fields["net_income_growth"] = v
			}
		}
		if _, ok := fields["pb_ratio"]; !ok {
			if v, ok := fields["pb"]; ok {
				fields["pb_ratio"] = v
			}
		}
		if _, ok := fields["pe_ratio"]; !ok {
			if v, ok := fields["pe"]; ok {
				fields["pe_ratio"] = v
			}
		}

		periods = append(periods, vnstock.FinancialPeriod{
			Symbol:  req.Symbol,
			Period:  req.Period,
			Year:    year,
			Quarter: quarter,
			Fields:  fields,
		})
	}

	// Sort by (Year, Quarter) descending — most recent first
	for i := 0; i < len(periods)-1; i++ {
		for j := i + 1; j < len(periods); j++ {
			if periods[i].Year < periods[j].Year ||
				(periods[i].Year == periods[j].Year && periods[i].Quarter < periods[j].Quarter) {
				periods[i], periods[j] = periods[j], periods[i]
			}
		}
	}

	return periods, nil
}

// toFloat64 safely converts an interface{} to float64.
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case json.Number:
		f, _ := val.Float64()
		return f
	default:
		return 0
	}
}

// FinancialRatios retrieves key financial ratios for a company.
// Uses VCI GraphQL API to fetch ratio data.
func (c *Connector) FinancialRatios(ctx context.Context, req vnstock.FinancialRatioRequest) (vnstock.FinancialRatio, error) {
	if req.Symbol == "" {
		return vnstock.FinancialRatio{}, &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: "symbol cannot be empty",
		}
	}

	// GraphQL query for financial ratios
	query := `
		query FinancialRatios($symbol: String!) {
			financialRatios(symbol: $symbol) {
				roa
				roe
				netProfitMargin
				revenueGrowth
				profitGrowth
				eps
				pe
				pb
				currentRatio
				debtToEquity
				dividendYield
				bookValuePerShare
				reportDate
			}
		}
	`

	variables := map[string]interface{}{
		"symbol": req.Symbol,
	}

	resp, err := c.doGraphQLRequest(ctx, query, variables)
	if err != nil {
		return vnstock.FinancialRatio{}, err
	}
	defer resp.Body.Close()

	// Parse GraphQL response
	var graphqlResp struct {
		Data struct {
			FinancialRatios struct {
				ROA               float64 `json:"roa"`
				ROE               float64 `json:"roe"`
				NetProfitMargin   float64 `json:"netProfitMargin"`
				RevenueGrowth     float64 `json:"revenueGrowth"`
				ProfitGrowth      float64 `json:"profitGrowth"`
				EPS               float64 `json:"eps"`
				PE                float64 `json:"pe"`
				PB                float64 `json:"pb"`
				CurrentRatio      float64 `json:"currentRatio"`
				DebtToEquity      float64 `json:"debtToEquity"`
				DividendYield     float64 `json:"dividendYield"`
				BookValuePerShare float64 `json:"bookValuePerShare"`
				ReportDate        string  `json:"reportDate"`
			} `json:"financialRatios"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&graphqlResp); err != nil {
		return vnstock.FinancialRatio{}, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode financial ratios response",
			Cause:   err,
		}
	}

	// Check for GraphQL errors
	if len(graphqlResp.Errors) > 0 {
		return vnstock.FinancialRatio{}, &vnstock.Error{
			Code:    vnstock.HTTPError,
			Message: fmt.Sprintf("GraphQL error: %s", graphqlResp.Errors[0].Message),
		}
	}

	data := graphqlResp.Data.FinancialRatios

	// Parse report date
	var reportDate time.Time
	if data.ReportDate != "" {
		reportDate, _ = time.Parse("2006-01-02", data.ReportDate)
	}

	return vnstock.FinancialRatio{
		Symbol:            req.Symbol,
		ReportDate:        reportDate,
		ROA:               data.ROA,
		ROE:               data.ROE,
		NetProfitMargin:   data.NetProfitMargin,
		RevenueGrowth:     data.RevenueGrowth,
		ProfitGrowth:      data.ProfitGrowth,
		EPS:               data.EPS,
		PE:                data.PE,
		PB:                data.PB,
		CurrentRatio:      data.CurrentRatio,
		DebtToEquity:      data.DebtToEquity,
		DividendYield:     data.DividendYield,
		BookValuePerShare: data.BookValuePerShare,
	}, nil
}
