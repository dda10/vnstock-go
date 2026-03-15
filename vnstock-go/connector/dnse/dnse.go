// Package dnse implements the Connector interface for the DNSE data source.
package dnse

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
	vnstock.RegisterConnector("DNSE", func(client *http.Client, logger *slog.Logger) vnstock.Connector {
		return New(client, logger)
	})
}

const (
	baseURL = "https://api.dnse.com.vn"
)

// Connector implements the vnstock.Connector interface for DNSE data source.
type Connector struct {
	client *http.Client
	logger *slog.Logger
}

// New creates a new DNSE connector with the provided HTTP client and logger.
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
	c.logger.Debug("DNSE request",
		slog.String("method", method),
		slog.String("url", url),
		slog.Int("status", statusCode),
		slog.Duration("elapsed", elapsed),
	)
}

// doRequest performs an HTTP request and logs it.
func (c *Connector) doRequest(ctx context.Context, method, endpoint string, params url.Values) (*http.Response, error) {
	reqURL := baseURL + endpoint
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.NetworkError,
			Message: "failed to create request",
			Cause:   err,
		}
	}

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

// QuoteHistory retrieves historical OHLCV data for a symbol.
func (c *Connector) QuoteHistory(ctx context.Context, req vnstock.QuoteHistoryRequest) ([]vnstock.Quote, error) {
	// Validate request
	if req.Start.After(req.End) {
		return nil, &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: "start date must be before or equal to end date",
		}
	}

	params := url.Values{}
	params.Set("symbol", req.Symbol)
	params.Set("start", req.Start.Format("2006-01-02"))
	params.Set("end", req.End.Format("2006-01-02"))
	params.Set("interval", req.Interval)

	resp, err := c.doRequest(ctx, http.MethodGet, "/api/quote/history", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Symbol    string  `json:"symbol"`
			Timestamp string  `json:"timestamp"`
			Open      float64 `json:"open"`
			High      float64 `json:"high"`
			Low       float64 `json:"low"`
			Close     float64 `json:"close"`
			Volume    int64   `json:"volume"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode response",
			Cause:   err,
		}
	}

	quotes := make([]vnstock.Quote, 0, len(result.Data))
	for _, item := range result.Data {
		ts, err := time.Parse(time.RFC3339, item.Timestamp)
		if err != nil {
			// Try alternative format
			ts, err = time.Parse("2006-01-02T15:04:05", item.Timestamp)
			if err != nil {
				return nil, &vnstock.Error{
					Code:    vnstock.SerialiseError,
					Message: "failed to parse timestamp",
					Cause:   err,
				}
			}
		}

		quotes = append(quotes, vnstock.Quote{
			Symbol:    item.Symbol,
			Timestamp: ts,
			Open:      item.Open,
			High:      item.High,
			Low:       item.Low,
			Close:     item.Close,
			Volume:    item.Volume,
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

	params := url.Values{}
	for _, symbol := range symbols {
		params.Add("symbols", symbol)
	}

	resp, err := c.doRequest(ctx, http.MethodGet, "/api/quote/realtime", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Symbol    string  `json:"symbol"`
			Timestamp string  `json:"timestamp"`
			Open      float64 `json:"open"`
			High      float64 `json:"high"`
			Low       float64 `json:"low"`
			Close     float64 `json:"close"`
			Volume    int64   `json:"volume"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode response",
			Cause:   err,
		}
	}

	quotes := make([]vnstock.Quote, 0, len(result.Data))
	for _, item := range result.Data {
		ts, err := time.Parse(time.RFC3339, item.Timestamp)
		if err != nil {
			ts, err = time.Parse("2006-01-02T15:04:05", item.Timestamp)
			if err != nil {
				return nil, &vnstock.Error{
					Code:    vnstock.SerialiseError,
					Message: "failed to parse timestamp",
					Cause:   err,
				}
			}
		}

		quotes = append(quotes, vnstock.Quote{
			Symbol:    item.Symbol,
			Timestamp: ts,
			Open:      item.Open,
			High:      item.High,
			Low:       item.Low,
			Close:     item.Close,
			Volume:    item.Volume,
			Interval:  "realtime",
		})
	}

	return quotes, nil
}

// Listing is not supported by DNSE.
func (c *Connector) Listing(ctx context.Context, exchange string) ([]vnstock.ListingRecord, error) {
	return nil, vnstock.ErrNotSupported
}

// IndexCurrent is not supported by DNSE.
func (c *Connector) IndexCurrent(ctx context.Context, name string) (vnstock.IndexRecord, error) {
	return vnstock.IndexRecord{}, vnstock.ErrNotSupported
}

// IndexHistory is not supported by DNSE.
func (c *Connector) IndexHistory(ctx context.Context, req vnstock.IndexHistoryRequest) ([]vnstock.IndexRecord, error) {
	return nil, vnstock.ErrNotSupported
}

// CompanyProfile retrieves descriptive information about a listed company.
func (c *Connector) CompanyProfile(ctx context.Context, symbol string) (vnstock.CompanyProfile, error) {
	params := url.Values{}
	params.Set("symbol", symbol)

	resp, err := c.doRequest(ctx, http.MethodGet, "/api/company/profile", params)
	if err != nil {
		return vnstock.CompanyProfile{}, err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Symbol      string `json:"symbol"`
			Name        string `json:"name"`
			Exchange    string `json:"exchange"`
			Sector      string `json:"sector"`
			Industry    string `json:"industry"`
			Founded     string `json:"founded"`
			Website     string `json:"website"`
			Description string `json:"description"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return vnstock.CompanyProfile{}, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode response",
			Cause:   err,
		}
	}

	return vnstock.CompanyProfile{
		Symbol:      result.Data.Symbol,
		Name:        result.Data.Name,
		Exchange:    result.Data.Exchange,
		Sector:      result.Data.Sector,
		Industry:    result.Data.Industry,
		Founded:     result.Data.Founded,
		Website:     result.Data.Website,
		Description: result.Data.Description,
	}, nil
}

// Officers retrieves the list of officers and executives for a company.
func (c *Connector) Officers(ctx context.Context, symbol string) ([]vnstock.Officer, error) {
	params := url.Values{}
	params.Set("symbol", symbol)

	resp, err := c.doRequest(ctx, http.MethodGet, "/api/company/officers", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Name            string `json:"name"`
			Title           string `json:"title"`
			AppointmentDate string `json:"appointment_date"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode response",
			Cause:   err,
		}
	}

	officers := make([]vnstock.Officer, 0, len(result.Data))
	for _, item := range result.Data {
		officers = append(officers, vnstock.Officer{
			Name:            item.Name,
			Title:           item.Title,
			AppointmentDate: item.AppointmentDate,
		})
	}

	return officers, nil
}

// FinancialStatement is not supported by DNSE.
func (c *Connector) FinancialStatement(ctx context.Context, req vnstock.FinancialRequest) ([]vnstock.FinancialPeriod, error) {
	return nil, vnstock.ErrNotSupported
}
