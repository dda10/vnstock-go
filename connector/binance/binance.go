// Package binance implements the Connector interface for the Binance data source.
package binance

import (
	"context"
	"log/slog"
	"net/http"

	vnstock "github.com/dda10/vnstock-go"
)

func init() {
	vnstock.RegisterConnector("Binance", func(client *http.Client, logger *slog.Logger) vnstock.Connector {
		return New(client, logger)
	})
}

// Connector implements the vnstock.Connector interface for Binance data source.
type Connector struct {
	client *http.Client
	logger *slog.Logger
}

// New creates a new Binance connector with the provided HTTP client and logger.
func New(client *http.Client, logger *slog.Logger) *Connector {
	if logger == nil {
		logger = slog.Default()
	}
	return &Connector{
		client: client,
		logger: logger,
	}
}

// QuoteHistory returns ErrNotSupported for Binance.
func (c *Connector) QuoteHistory(ctx context.Context, req vnstock.QuoteHistoryRequest) ([]vnstock.Quote, error) {
	return nil, vnstock.ErrNotSupported
}

// RealTimeQuotes returns ErrNotSupported for Binance.
func (c *Connector) RealTimeQuotes(ctx context.Context, symbols []string) ([]vnstock.Quote, error) {
	return nil, vnstock.ErrNotSupported
}

// Listing returns ErrNotSupported for Binance.
func (c *Connector) Listing(ctx context.Context, exchange string) ([]vnstock.ListingRecord, error) {
	return nil, vnstock.ErrNotSupported
}

// IndexCurrent returns ErrNotSupported for Binance.
func (c *Connector) IndexCurrent(ctx context.Context, name string) (vnstock.IndexRecord, error) {
	return vnstock.IndexRecord{}, vnstock.ErrNotSupported
}

// IndexHistory returns ErrNotSupported for Binance.
func (c *Connector) IndexHistory(ctx context.Context, req vnstock.IndexHistoryRequest) ([]vnstock.IndexRecord, error) {
	return nil, vnstock.ErrNotSupported
}

// CompanyProfile returns ErrNotSupported for Binance (crypto doesn't have company profiles).
func (c *Connector) CompanyProfile(ctx context.Context, symbol string) (vnstock.CompanyProfile, error) {
	return vnstock.CompanyProfile{}, vnstock.ErrNotSupported
}

// Officers returns ErrNotSupported for Binance (crypto doesn't have officers).
func (c *Connector) Officers(ctx context.Context, symbol string) ([]vnstock.Officer, error) {
	return nil, vnstock.ErrNotSupported
}

// FinancialStatement returns ErrNotSupported for Binance (crypto doesn't have financial statements).
func (c *Connector) FinancialStatement(ctx context.Context, req vnstock.FinancialRequest) ([]vnstock.FinancialPeriod, error) {
	return nil, vnstock.ErrNotSupported
}

// FinancialRatios returns ErrNotSupported for Binance (crypto doesn't have financial ratios).
func (c *Connector) FinancialRatios(ctx context.Context, req vnstock.FinancialRatioRequest) (vnstock.FinancialRatio, error) {
	return vnstock.FinancialRatio{}, vnstock.ErrNotSupported
}
