package vnstock

import "context"

// Connector is the interface that all data source adapters must implement.
// Each method corresponds to a data retrieval operation. Connectors that
// do not support a particular method should return ErrNotSupported.
type Connector interface {
	// QuoteHistory retrieves historical OHLCV data for a symbol.
	QuoteHistory(ctx context.Context, req QuoteHistoryRequest) ([]Quote, error)

	// RealTimeQuotes retrieves the most recent quote for one or more symbols.
	RealTimeQuotes(ctx context.Context, symbols []string) ([]Quote, error)

	// Listing retrieves the full list of symbols traded on an exchange.
	// If exchange is empty, returns all symbols across all exchanges.
	Listing(ctx context.Context, exchange string) ([]ListingRecord, error)

	// IndexCurrent retrieves the current value of a named market index.
	IndexCurrent(ctx context.Context, name string) (IndexRecord, error)

	// IndexHistory retrieves historical values for a named market index.
	IndexHistory(ctx context.Context, req IndexHistoryRequest) ([]IndexRecord, error)

	// CompanyProfile retrieves descriptive information about a listed company.
	CompanyProfile(ctx context.Context, symbol string) (CompanyProfile, error)

	// Officers retrieves the list of officers and executives for a company.
	Officers(ctx context.Context, symbol string) ([]Officer, error)

	// FinancialStatement retrieves financial statement data for a company.
	FinancialStatement(ctx context.Context, req FinancialRequest) ([]FinancialPeriod, error)
}

// ErrNotSupported is a sentinel error returned by connectors that do not
// support a particular method.
var ErrNotSupported = &Error{
	Code:    NotSupported,
	Message: "operation not supported by this connector",
}

// MockConnector is a stub implementation of the Connector interface that
// returns ErrNotSupported for all methods. It is intended for use in tests.
type MockConnector struct{}

// QuoteHistory returns ErrNotSupported.
func (m *MockConnector) QuoteHistory(ctx context.Context, req QuoteHistoryRequest) ([]Quote, error) {
	return nil, ErrNotSupported
}

// RealTimeQuotes returns ErrNotSupported.
func (m *MockConnector) RealTimeQuotes(ctx context.Context, symbols []string) ([]Quote, error) {
	return nil, ErrNotSupported
}

// Listing returns ErrNotSupported.
func (m *MockConnector) Listing(ctx context.Context, exchange string) ([]ListingRecord, error) {
	return nil, ErrNotSupported
}

// IndexCurrent returns ErrNotSupported.
func (m *MockConnector) IndexCurrent(ctx context.Context, name string) (IndexRecord, error) {
	return IndexRecord{}, ErrNotSupported
}

// IndexHistory returns ErrNotSupported.
func (m *MockConnector) IndexHistory(ctx context.Context, req IndexHistoryRequest) ([]IndexRecord, error) {
	return nil, ErrNotSupported
}

// CompanyProfile returns ErrNotSupported.
func (m *MockConnector) CompanyProfile(ctx context.Context, symbol string) (CompanyProfile, error) {
	return CompanyProfile{}, ErrNotSupported
}

// Officers returns ErrNotSupported.
func (m *MockConnector) Officers(ctx context.Context, symbol string) ([]Officer, error) {
	return nil, ErrNotSupported
}

// FinancialStatement returns ErrNotSupported.
func (m *MockConnector) FinancialStatement(ctx context.Context, req FinancialRequest) ([]FinancialPeriod, error) {
	return nil, ErrNotSupported
}
