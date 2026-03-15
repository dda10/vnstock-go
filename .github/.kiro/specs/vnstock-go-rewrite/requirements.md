# Requirements Document

## Introduction

vnstock-go is a Go rewrite of the vnstock Python library, providing programmatic access to Vietnamese stock market data. The library exposes stock quotes, company financials, market listings, indices, and multi-source data connectors through an idiomatic Go API. The rewrite targets feature parity with the Python version while leveraging Go's strengths: static typing, native concurrency, and high performance.

## Glossary

- **Library**: The vnstock-go Go module being specified
- **Client**: A caller of the Library (application code or end user)
- **Connector**: A data source adapter that fetches data from a specific upstream provider (e.g., VCI, DNSE, FMP, Binance)
- **Symbol**: A stock ticker string identifying a listed security (e.g., "VNM", "HPG")
- **Quote**: A price record for a Symbol at a point in time, including open, high, low, close, and volume
- **OHLCV**: Open, High, Low, Close, Volume — the standard fields of a Quote
- **Interval**: A candlestick time resolution (e.g., 1m, 5m, 15m, 1h, 1d, 1w, 1M)
- **Listing**: The set of all Symbols currently traded on a Vietnamese exchange
- **Index**: A market index such as VN-Index, HNX-Index, or UPCOM-Index
- **Financial_Statement**: Structured financial data for a company (income statement, balance sheet, cash flow)
- **Company_Profile**: Non-financial descriptive data about a listed company (name, sector, officers, etc.)
- **Proxy**: An HTTP/HTTPS proxy server used to route outbound requests from the Library
- **Exporter**: A component that serialises Library data into an external format (CSV, JSON, Excel)
- **Serialiser**: A component that converts between Go structs and wire formats (JSON, CSV)
- **Pretty_Printer**: A component that formats Go structs back into their canonical wire representation

---

## Requirements

### Requirement 1: Stock Quote History

**User Story:** As a developer, I want to retrieve historical OHLCV data for a stock symbol, so that I can perform backtesting and technical analysis.

#### Acceptance Criteria

1. WHEN a Client requests historical quotes for a valid Symbol with a start date, end date, and Interval, THE Library SHALL return a slice of Quote records ordered by ascending timestamp.
2. WHEN a Client requests historical quotes and the upstream Connector returns an error, THE Library SHALL return a descriptive error wrapping the upstream cause.
3. WHEN a Client requests historical quotes with a start date after the end date, THE Library SHALL return an error before making any network request.
4. THE Quote SHALL contain Symbol, Timestamp, Open, High, Low, Close, Volume, and Interval fields with their correct Go types.
5. WHILE a historical quote request is in progress, THE Library SHALL respect any configured request timeout and return a timeout error if the deadline is exceeded.

---

### Requirement 2: Real-Time Quote

**User Story:** As a developer, I want to fetch the latest price quote for one or more symbols, so that I can display live market data in my application.

#### Acceptance Criteria

1. WHEN a Client requests real-time quotes for a list of one or more Symbols, THE Library SHALL return the most recent Quote for each requested Symbol.
2. WHEN a Client requests real-time quotes and a Symbol is not found by the Connector, THE Library SHALL return an error identifying the missing Symbol.
3. WHEN a Client requests real-time quotes for multiple Symbols concurrently, THE Library SHALL issue upstream requests concurrently and aggregate results without data races.

---

### Requirement 3: Market Listing

**User Story:** As a developer, I want to retrieve the full list of symbols traded on Vietnamese exchanges, so that I can build screeners and portfolio tools.

#### Acceptance Criteria

1. WHEN a Client requests the market Listing, THE Library SHALL return a slice of Listing records containing Symbol, exchange, company name, and sector for every currently traded security.
2. WHEN a Client filters the Listing by exchange (HOSE, HNX, UPCOM), THE Library SHALL return only records matching the specified exchange.
3. IF the upstream Connector returns an empty Listing, THEN THE Library SHALL return an error indicating that no listing data was available.

---

### Requirement 4: Market Indices

**User Story:** As a developer, I want to retrieve current and historical values for Vietnamese market indices, so that I can track overall market performance.

#### Acceptance Criteria

1. WHEN a Client requests the current value of a named Index, THE Library SHALL return the index value, change, percentage change, and timestamp.
2. WHEN a Client requests historical values for a named Index with a start date, end date, and Interval, THE Library SHALL return a slice of index OHLCV records ordered by ascending timestamp.
3. IF a Client requests an Index name that is not supported by the active Connector, THEN THE Library SHALL return an error listing the supported index names.

---

### Requirement 5: Company Profile

**User Story:** As a developer, I want to retrieve descriptive information about a listed company, so that I can display company details in my application.

#### Acceptance Criteria

1. WHEN a Client requests a Company_Profile for a valid Symbol, THE Library SHALL return the company name, exchange, sector, industry, founding date, website, and a brief description.
2. WHEN a Client requests a Company_Profile for a Symbol that is not listed, THE Library SHALL return an error identifying the unknown Symbol.
3. WHEN a Client requests officer information for a valid Symbol, THE Library SHALL return a slice of officer records each containing name, title, and appointment date.

---

### Requirement 6: Financial Statements

**User Story:** As a developer, I want to retrieve income statements, balance sheets, and cash flow statements for a listed company, so that I can perform fundamental analysis.

#### Acceptance Criteria

1. WHEN a Client requests a Financial_Statement of a specified type (income, balance, cashflow) for a valid Symbol and period (annual or quarterly), THE Library SHALL return a slice of period records in descending chronological order.
2. WHEN a Client requests a Financial_Statement and the Connector returns incomplete data, THE Library SHALL return the partial data alongside a warning indicating which fields are missing.
3. IF a Client requests a Financial_Statement for a Symbol that has no financial data available, THEN THE Library SHALL return an error distinguishing "symbol not found" from "no financial data available".

---

### Requirement 7: Multi-Source Connector Architecture

**User Story:** As a developer, I want to switch between data sources (VCI, DNSE, FMP, Binance) without changing my application code, so that I can choose the best source for each use case.

#### Acceptance Criteria

1. THE Library SHALL define a Connector interface that all data source adapters implement, covering methods for quotes, listings, indices, company profiles, and financial statements.
2. WHEN a Client instantiates the Library with a named Connector (e.g., "VCI", "DNSE", "FMP", "Binance"), THE Library SHALL route all data requests through that Connector.
3. IF a Client specifies an unrecognised Connector name, THEN THE Library SHALL return an error at construction time listing the available Connector names.
4. WHERE a Connector does not support a particular data method, THE Library SHALL return an explicit "not supported" error for that method rather than panicking.

---

### Requirement 8: Proxy Support

**User Story:** As a developer, I want to route Library requests through an HTTP proxy, so that I can operate in restricted network environments.

#### Acceptance Criteria

1. WHERE a Proxy URL is provided in the Library configuration, THE Library SHALL route all outbound HTTP requests through the specified Proxy.
2. WHEN a Proxy is configured and the Proxy server is unreachable, THE Library SHALL return an error identifying the Proxy as the failure point.
3. THE Library SHALL support HTTP and HTTPS Proxy URLs in the standard `http://host:port` and `https://host:port` formats.
4. WHERE no Proxy is configured, THE Library SHALL use the default system HTTP transport.

---

### Requirement 9: Data Export

**User Story:** As a developer, I want to export Library data to CSV, JSON, and Excel formats, so that I can share results with non-technical stakeholders.

#### Acceptance Criteria

1. WHEN a Client exports a slice of Quote records to CSV, THE Exporter SHALL write a valid CSV file with a header row and one data row per Quote.
2. WHEN a Client exports a slice of Quote records to JSON, THE Exporter SHALL write a valid JSON array where each element represents one Quote.
3. WHERE Excel export is enabled, WHEN a Client exports data to Excel, THE Exporter SHALL write a valid `.xlsx` file with one worksheet per data type.
4. IF the target file path is not writable, THEN THE Exporter SHALL return an error identifying the path and the reason for the failure.
5. THE Exporter SHALL support exporting all primary data types: Quote, Listing, Index values, Company_Profile, and Financial_Statement records.

---

### Requirement 10: JSON Serialisation Round-Trip

**User Story:** As a developer, I want all Library data types to serialise and deserialise correctly, so that I can store and transmit data reliably.

#### Acceptance Criteria

1. THE Serialiser SHALL serialise every primary data struct (Quote, Listing record, Index record, Company_Profile, Financial_Statement record) to valid JSON.
2. THE Pretty_Printer SHALL format serialised JSON with consistent field ordering and indentation.
3. FOR ALL valid primary data structs, serialising then deserialising SHALL produce a struct equal to the original (round-trip property).
4. WHEN the Serialiser receives a JSON payload with an unrecognised field, THE Serialiser SHALL ignore the unknown field and return the partially populated struct without error.
5. WHEN the Serialiser receives a JSON payload with a missing required field, THE Serialiser SHALL return a descriptive error identifying the missing field.

---

### Requirement 11: Concurrency Safety

**User Story:** As a developer, I want to use the Library safely from multiple goroutines, so that I can build high-throughput data pipelines.

#### Acceptance Criteria

1. THE Library SHALL be safe to call from multiple goroutines concurrently without external synchronisation by the Client.
2. WHEN multiple goroutines request data for different Symbols simultaneously, THE Library SHALL process each request independently without shared mutable state between requests.
3. WHERE a connection pool is used by a Connector, THE Library SHALL manage pool lifecycle internally and not require the Client to manage connections.

---

### Requirement 12: Configuration and Initialisation

**User Story:** As a developer, I want a clear and type-safe way to configure the Library, so that I can set timeouts, proxies, and data sources without runtime surprises.

#### Acceptance Criteria

1. THE Library SHALL expose a `Config` struct with fields for Connector name, Proxy URL, request timeout, and maximum retry count.
2. WHEN a Client creates a Library instance with a zero-value `Config`, THE Library SHALL apply documented default values for all fields.
3. THE Library SHALL provide a constructor function that validates the `Config` and returns an error for any invalid field value before any network activity occurs.
4. WHERE environment variables are set for Connector name or Proxy URL, THE Library SHALL read those values as defaults when the corresponding `Config` fields are empty.

---

### Requirement 13: Error Handling

**User Story:** As a developer, I want structured, inspectable errors from the Library, so that I can handle different failure modes programmatically.

#### Acceptance Criteria

1. THE Library SHALL define a typed `Error` type that carries an error code, a human-readable message, and an optional wrapped cause.
2. WHEN a network error occurs, THE Library SHALL wrap the underlying error in a Library `Error` with a `NetworkError` code.
3. WHEN an upstream Connector returns an HTTP status code outside the 2xx range, THE Library SHALL return a Library `Error` carrying the HTTP status code.
4. THE Library `Error` type SHALL implement the standard Go `error` interface and be unwrappable via `errors.As` and `errors.Is`.

---

### Requirement 14: Logging and Observability

**User Story:** As a developer, I want structured log output from the Library, so that I can diagnose issues in production without modifying Library code.

#### Acceptance Criteria

1. THE Library SHALL emit structured log entries using the standard `log/slog` package.
2. WHEN a Client provides a custom `slog.Logger` in the `Config`, THE Library SHALL use that logger for all log output.
3. WHERE no custom logger is provided, THE Library SHALL use the default `slog` logger.
4. THE Library SHALL log outbound request URL, HTTP method, response status code, and elapsed time at the DEBUG level for every Connector request.
