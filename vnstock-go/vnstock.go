// Package vnstock provides programmatic access to Vietnamese stock market data.
package vnstock

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/dda10/vnstock-go/internal/httpclient"
)

// validConnectors is the set of recognised connector names.
var validConnectors = map[string]bool{
	"VCI":     true,
	"DNSE":    true,
	"FMP":     true,
	"Binance": true,
	"KBS":     true,
	"GOLD":    true,
}

// Config holds configuration options for the vnstock Client.
type Config struct {
	// Connector is the name of the data source connector ("VCI", "DNSE", "FMP", "Binance", "KBS").
	// If empty, the VNSTOCK_CONNECTOR environment variable is used.
	Connector string

	// ProxyURL is the HTTP/HTTPS proxy URL (e.g., "http://host:port").
	// If empty, the VNSTOCK_PROXY_URL environment variable is used.
	// If neither is set, no proxy is used.
	ProxyURL string

	// Timeout is the request timeout duration. Default: 30s.
	Timeout time.Duration

	// MaxRetries is the maximum number of retry attempts. Default: 3.
	MaxRetries int

	// Logger is the structured logger for debug output. Default: slog.Default().
	Logger *slog.Logger
}

// Client is the main entry point for accessing Vietnamese stock market data.
type Client struct {
	connector  Connector
	logger     *slog.Logger
	config     Config
	httpClient *http.Client
}

// New creates a new Client with the given configuration.
// It applies environment variable defaults, validates all fields, and constructs
// the named connector. Returns *Error{Code: ConfigError} for any invalid field.
func New(cfg Config) (*Client, error) {
	// Apply environment variable defaults
	if cfg.Connector == "" {
		cfg.Connector = os.Getenv("VNSTOCK_CONNECTOR")
	}
	if cfg.ProxyURL == "" {
		cfg.ProxyURL = os.Getenv("VNSTOCK_PROXY_URL")
	}

	// Apply field defaults
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	// Validate connector name
	if cfg.Connector == "" {
		return nil, &Error{
			Code:    ConfigError,
			Message: "connector name is required (set Config.Connector or VNSTOCK_CONNECTOR env var)",
		}
	}
	if !validConnectors[cfg.Connector] {
		return nil, &Error{
			Code:    ConfigError,
			Message: "unrecognised connector name: " + cfg.Connector + "; valid names are: VCI, DNSE, FMP, Binance, KBS, GOLD",
		}
	}

	// Validate timeout
	if cfg.Timeout < 0 {
		return nil, &Error{
			Code:    ConfigError,
			Message: "timeout must not be negative",
		}
	}

	// Validate proxy URL by attempting to create an HTTP client
	// This also validates the proxy URL format
	httpClient, err := httpclient.New(cfg.ProxyURL, cfg.Timeout)
	if err != nil {
		// Wrap httpclient.ProxyError as vnstock.Error with ConfigError code
		return nil, &Error{
			Code:    ConfigError,
			Message: err.Error(),
			Cause:   err,
		}
	}

	// Construct the named connector
	connector, err := newConnector(cfg.Connector, httpClient, cfg.Logger)
	if err != nil {
		return nil, err
	}

	return &Client{
		connector:  connector,
		logger:     cfg.Logger,
		config:     cfg,
		httpClient: httpClient,
	}, nil
}

// newConnector creates a connector instance by name using the registry.
func newConnector(name string, httpClient *http.Client, logger *slog.Logger) (Connector, error) {
	factory, ok := getConnectorFactory(name)
	if !ok {
		return nil, &Error{
			Code:    ConfigError,
			Message: "unrecognised connector name: " + name + "; valid names are: VCI, DNSE, FMP, Binance, KBS, GOLD",
		}
	}
	return factory(httpClient, logger), nil
}

// QuoteHistory retrieves historical OHLCV data for a symbol.
// It validates that Start < End before making any network call.
// Returns *Error with InvalidInput code if validation fails.
func (c *Client) QuoteHistory(ctx context.Context, req QuoteHistoryRequest) ([]Quote, error) {
	// Validate Start < End
	if !req.Start.Before(req.End) {
		return nil, &Error{
			Code:    InvalidInput,
			Message: "start date must be before end date",
		}
	}

	// Delegate to connector
	quotes, err := c.connector.QuoteHistory(ctx, req)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return quotes, nil
}

// RealTimeQuotes retrieves the most recent quote for one or more symbols.
// It uses goroutines to fetch quotes concurrently and aggregates results
// without shared mutable state.
func (c *Client) RealTimeQuotes(ctx context.Context, symbols []string) ([]Quote, error) {
	if len(symbols) == 0 {
		return []Quote{}, nil
	}

	// Use a channel to collect results
	type result struct {
		quotes []Quote
		err    error
	}
	results := make(chan result, len(symbols))

	// Launch a goroutine for each symbol
	for _, symbol := range symbols {
		go func(sym string) {
			quotes, err := c.connector.RealTimeQuotes(ctx, []string{sym})
			results <- result{quotes: quotes, err: err}
		}(symbol)
	}

	// Collect results
	var allQuotes []Quote
	var firstErr error
	for i := 0; i < len(symbols); i++ {
		res := <-results
		if res.err != nil && firstErr == nil {
			firstErr = res.err
		}
		if res.quotes != nil {
			allQuotes = append(allQuotes, res.quotes...)
		}
	}

	// If any error occurred, wrap and return it
	if firstErr != nil {
		return nil, c.wrapError(firstErr)
	}

	return allQuotes, nil
}

// Listing retrieves the full list of symbols traded on an exchange.
// If exchange is empty, returns all symbols across all exchanges.
func (c *Client) Listing(ctx context.Context, exchange string) ([]ListingRecord, error) {
	records, err := c.connector.Listing(ctx, exchange)
	if err != nil {
		return nil, c.wrapError(err)
	}
	return records, nil
}

// IndexCurrent retrieves the current value of a named market index.
func (c *Client) IndexCurrent(ctx context.Context, name string) (IndexRecord, error) {
	record, err := c.connector.IndexCurrent(ctx, name)
	if err != nil {
		return IndexRecord{}, c.wrapError(err)
	}
	return record, nil
}

// IndexHistory retrieves historical values for a named market index.
// It validates that Start < End before making any network call.
func (c *Client) IndexHistory(ctx context.Context, req IndexHistoryRequest) ([]IndexRecord, error) {
	// Validate Start < End
	if !req.Start.Before(req.End) {
		return nil, &Error{
			Code:    InvalidInput,
			Message: "start date must be before end date",
		}
	}

	records, err := c.connector.IndexHistory(ctx, req)
	if err != nil {
		return nil, c.wrapError(err)
	}
	return records, nil
}

// CompanyProfile retrieves descriptive information about a listed company.
func (c *Client) CompanyProfile(ctx context.Context, symbol string) (CompanyProfile, error) {
	profile, err := c.connector.CompanyProfile(ctx, symbol)
	if err != nil {
		return CompanyProfile{}, c.wrapError(err)
	}
	return profile, nil
}

// Officers retrieves the list of officers and executives for a company.
func (c *Client) Officers(ctx context.Context, symbol string) ([]Officer, error) {
	officers, err := c.connector.Officers(ctx, symbol)
	if err != nil {
		return nil, c.wrapError(err)
	}
	return officers, nil
}

// FinancialStatement retrieves financial statement data for a company.
func (c *Client) FinancialStatement(ctx context.Context, req FinancialRequest) ([]FinancialPeriod, error) {
	periods, err := c.connector.FinancialStatement(ctx, req)
	if err != nil {
		return nil, c.wrapError(err)
	}
	return periods, nil
}

// GoldPrice retrieves gold prices from various sources (SJC, BTMC).
// This method uses a dedicated GOLD connector, separate from stock market connectors.
func (c *Client) GoldPrice(ctx context.Context, req GoldPriceRequest) ([]GoldPrice, error) {
	// Create a temporary GOLD connector client
	goldClient, err := New(Config{
		Connector:  "GOLD",
		Timeout:    c.config.Timeout,
		MaxRetries: c.config.MaxRetries,
		Logger:     c.logger,
	})
	if err != nil {
		return nil, &Error{
			Code:    ConfigError,
			Message: "failed to create GOLD connector",
			Cause:   err,
		}
	}

	// The GOLD connector implements a special GoldPrice method
	type GoldConnector interface {
		GoldPrice(ctx context.Context, req GoldPriceRequest) ([]GoldPrice, error)
	}

	if goldConn, ok := goldClient.connector.(GoldConnector); ok {
		prices, err := goldConn.GoldPrice(ctx, req)
		if err != nil {
			return nil, c.wrapError(err)
		}
		return prices, nil
	}

	return nil, &Error{
		Code:    NotSupported,
		Message: "GOLD connector does not support GoldPrice method",
	}
}

// wrapError wraps connector errors as *Error with appropriate error codes.
// If the error is already a *Error, it is returned as-is.
// Network errors are wrapped as NetworkError, HTTP errors as HTTPError.
func (c *Client) wrapError(err error) error {
	if err == nil {
		return nil
	}

	// If already a *Error, return as-is
	if _, ok := err.(*Error); ok {
		return err
	}

	// For now, return a generic network error
	// This will be enhanced when we have actual connector implementations
	// that can distinguish between network and HTTP errors
	return &Error{
		Code:    NetworkError,
		Message: "connector operation failed",
		Cause:   err,
	}
}
