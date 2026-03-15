// Package gold provides a connector for retrieving gold prices from Vietnamese sources.
package gold

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	vnstock "github.com/dda10/vnstock-go"
)

// Connector implements the vnstock.Connector interface for gold price data.
type Connector struct {
	client *http.Client
	logger *slog.Logger
}

func init() {
	vnstock.RegisterConnector("GOLD", New)
}

// New creates a new gold price connector.
func New(client *http.Client, logger *slog.Logger) vnstock.Connector {
	return &Connector{
		client: client,
		logger: logger,
	}
}

// logRequest logs HTTP request details.
func (c *Connector) logRequest(method, url string, statusCode int, elapsed time.Duration) {
	c.logger.Debug("gold connector request",
		"method", method,
		"url", url,
		"status", statusCode,
		"elapsed_ms", elapsed.Milliseconds(),
	)
}

// GoldPrice retrieves gold prices from SJC and/or BTMC sources.
func (c *Connector) GoldPrice(ctx context.Context, req vnstock.GoldPriceRequest) ([]vnstock.GoldPrice, error) {
	var allPrices []vnstock.GoldPrice

	// Default to today if no date specified
	date := req.Date
	if date.IsZero() {
		date = time.Now()
	}

	// Fetch from requested source(s)
	switch req.Source {
	case "SJC":
		prices, err := c.fetchSJCPrices(ctx, date)
		if err != nil {
			return nil, err
		}
		allPrices = append(allPrices, prices...)
	case "BTMC":
		prices, err := c.fetchBTMCPrices(ctx)
		if err != nil {
			return nil, err
		}
		allPrices = append(allPrices, prices...)
	default:
		// Fetch from all sources
		sjcPrices, sjcErr := c.fetchSJCPrices(ctx, date)
		if sjcErr != nil {
			c.logger.Warn("failed to fetch SJC prices", "error", sjcErr)
		} else {
			allPrices = append(allPrices, sjcPrices...)
		}

		btmcPrices, btmcErr := c.fetchBTMCPrices(ctx)
		if btmcErr != nil {
			c.logger.Warn("failed to fetch BTMC prices", "error", btmcErr)
		} else {
			allPrices = append(allPrices, btmcPrices...)
		}

		// If both sources failed, return error
		if sjcErr != nil && btmcErr != nil {
			return nil, &vnstock.Error{
				Code:    vnstock.NetworkError,
				Message: "failed to fetch gold prices from all sources",
				Cause:   sjcErr,
			}
		}
	}

	if len(allPrices) == 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "no gold price data available",
		}
	}

	return allPrices, nil
}

// fetchSJCPrices retrieves gold prices from SJC API.
func (c *Connector) fetchSJCPrices(ctx context.Context, date time.Time) ([]vnstock.GoldPrice, error) {
	apiURL := "https://sjc.com.vn/GoldPrice/Services/PriceService.ashx"

	// Format date as DD/MM/YYYY
	formattedDate := date.Format("02/01/2006")

	// Prepare POST payload
	payload := fmt.Sprintf("method=GetSJCGoldPriceByDate&toDate=%s", formattedDate)

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBufferString(payload))
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.NetworkError,
			Message: "failed to create SJC request",
			Cause:   err,
		}
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := c.client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		c.logRequest("POST", apiURL, 0, elapsed)
		return nil, &vnstock.Error{
			Code:    vnstock.NetworkError,
			Message: "SJC API request failed",
			Cause:   err,
		}
	}
	defer resp.Body.Close()

	c.logRequest("POST", apiURL, resp.StatusCode, elapsed)

	if resp.StatusCode != http.StatusOK {
		return nil, &vnstock.Error{
			Code:       vnstock.HTTPError,
			Message:    "SJC API returned non-200 status",
			StatusCode: resp.StatusCode,
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.NetworkError,
			Message: "failed to read SJC response body",
			Cause:   err,
		}
	}

	// Parse JSON response
	var apiResp struct {
		Success bool `json:"success"`
		Data    []struct {
			TypeName   string  `json:"TypeName"`
			BranchName string  `json:"BranchName"`
			BuyValue   float64 `json:"BuyValue"`
			SellValue  float64 `json:"SellValue"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse SJC JSON response",
			Cause:   err,
		}
	}

	if !apiResp.Success {
		return nil, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "SJC API returned success=false",
		}
	}

	if len(apiResp.Data) == 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "SJC API returned no data",
		}
	}

	// Convert to GoldPrice records
	prices := make([]vnstock.GoldPrice, 0, len(apiResp.Data))
	for _, item := range apiResp.Data {
		prices = append(prices, vnstock.GoldPrice{
			TypeName:  item.TypeName,
			Branch:    item.BranchName,
			BuyPrice:  item.BuyValue,
			SellPrice: item.SellValue,
			Date:      date,
			Source:    "SJC",
		})
	}

	return prices, nil
}

// fetchBTMCPrices retrieves gold prices from Bảo Tín Minh Châu API.
func (c *Connector) fetchBTMCPrices(ctx context.Context) ([]vnstock.GoldPrice, error) {
	apiURL := "http://api.btmc.vn/api/BTMCAPI/getpricebtmc?key=3kd8ub1llcg9t45hnoh8hmn7t5kc2v"

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.NetworkError,
			Message: "failed to create BTMC request",
			Cause:   err,
		}
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := c.client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		c.logRequest("GET", apiURL, 0, elapsed)
		return nil, &vnstock.Error{
			Code:    vnstock.NetworkError,
			Message: "BTMC API request failed",
			Cause:   err,
		}
	}
	defer resp.Body.Close()

	c.logRequest("GET", apiURL, resp.StatusCode, elapsed)

	if resp.StatusCode != http.StatusOK {
		return nil, &vnstock.Error{
			Code:       vnstock.HTTPError,
			Message:    "BTMC API returned non-200 status",
			StatusCode: resp.StatusCode,
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.NetworkError,
			Message: "failed to read BTMC response body",
			Cause:   err,
		}
	}

	// Parse JSON response
	var apiResp struct {
		DataList struct {
			Data []map[string]interface{} `json:"Data"`
		} `json:"DataList"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse BTMC JSON response",
			Cause:   err,
		}
	}

	if len(apiResp.DataList.Data) == 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "BTMC API returned no data",
		}
	}

	// Convert to GoldPrice records
	prices := make([]vnstock.GoldPrice, 0, len(apiResp.DataList.Data))
	now := time.Now()

	for _, item := range apiResp.DataList.Data {
		rowNum, ok := item["@row"].(float64)
		if !ok {
			continue
		}
		row := fmt.Sprintf("%.0f", rowNum)

		// Extract fields using row number
		typeName := getStringField(item, fmt.Sprintf("@n_%s", row))
		buyPrice := getFloatField(item, fmt.Sprintf("@pb_%s", row))
		sellPrice := getFloatField(item, fmt.Sprintf("@ps_%s", row))

		if typeName != "" {
			prices = append(prices, vnstock.GoldPrice{
				TypeName:  typeName,
				Branch:    "",
				BuyPrice:  buyPrice,
				SellPrice: sellPrice,
				Date:      now,
				Source:    "BTMC",
			})
		}
	}

	return prices, nil
}

// Helper functions for parsing BTMC response
func getStringField(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getFloatField(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case string:
			// Try to parse string as float
			var f float64
			fmt.Sscanf(v, "%f", &f)
			return f
		}
	}
	return 0
}

// Stub implementations for other Connector interface methods
func (c *Connector) QuoteHistory(ctx context.Context, req vnstock.QuoteHistoryRequest) ([]vnstock.Quote, error) {
	return nil, vnstock.ErrNotSupported
}

func (c *Connector) RealTimeQuotes(ctx context.Context, symbols []string) ([]vnstock.Quote, error) {
	return nil, vnstock.ErrNotSupported
}

func (c *Connector) Listing(ctx context.Context, exchange string) ([]vnstock.ListingRecord, error) {
	return nil, vnstock.ErrNotSupported
}

func (c *Connector) IndexCurrent(ctx context.Context, name string) (vnstock.IndexRecord, error) {
	return vnstock.IndexRecord{}, vnstock.ErrNotSupported
}

func (c *Connector) IndexHistory(ctx context.Context, req vnstock.IndexHistoryRequest) ([]vnstock.IndexRecord, error) {
	return nil, vnstock.ErrNotSupported
}

func (c *Connector) CompanyProfile(ctx context.Context, symbol string) (vnstock.CompanyProfile, error) {
	return vnstock.CompanyProfile{}, vnstock.ErrNotSupported
}

func (c *Connector) Officers(ctx context.Context, symbol string) ([]vnstock.Officer, error) {
	return nil, vnstock.ErrNotSupported
}

func (c *Connector) FinancialStatement(ctx context.Context, req vnstock.FinancialRequest) ([]vnstock.FinancialPeriod, error) {
	return nil, vnstock.ErrNotSupported
}
