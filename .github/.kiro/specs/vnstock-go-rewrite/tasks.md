# Implementation Plan: vnstock-go

## Overview

Incremental Go module build: scaffold → core types → errors → config/client → HTTP factory → connector interface → connector implementations → exporters → logging → wire everything together. Each step compiles and is tested before the next begins.

## Tasks

- [x] 1. Scaffold module and package structure
  - Create `go.mod` for `github.com/user/vnstock-go` with Go 1.22+
  - Add dependencies: `pgregory.net/rapid`, `github.com/xuri/excelize/v2`
  - Create empty package files: `vnstock.go`, `connector.go`, `errors.go`, `models.go`
  - Create subdirectory stubs: `exporter/`, `connector/vci/`, `connector/dnse/`, `connector/fmp/`, `connector/binance/`, `internal/httpclient/`
  - _Requirements: 7.1, 12.1_

- [x] 2. Define data models
  - [x] 2.1 Implement all primary structs in `models.go`
    - `Quote`, `ListingRecord`, `IndexRecord`, `CompanyProfile`, `Officer`, `FinancialPeriod`
    - Request types: `QuoteHistoryRequest`, `IndexHistoryRequest`, `FinancialRequest`
    - Add `json` struct tags on all fields; `time.Time` fields use RFC 3339 via `time.Time` default marshalling
    - _Requirements: 1.4, 3.1, 4.1, 5.1, 5.3, 6.1, 10.1_

  - [x] 2.2 Write property test for JSON round-trip (Property 16)
    - Property 16: JSON serialisation round-trip
    - Use `rapid` generators for each primary struct
    - Marshal → Unmarshal → assert equality for `Quote`, `ListingRecord`, `IndexRecord`, `CompanyProfile`, `FinancialPeriod`
    - _Validates: Requirements 10.1, 10.3_

  - [x] 2.3 Write unit tests for unknown field handling and missing required field error (Properties 17)
    - Property 17: Unknown JSON fields are silently ignored
    - Test that extra JSON keys do not cause unmarshal errors
    - Test that a missing required field returns a descriptive error
    - _Validates: Requirements 10.4, 10.5_

- [x] 3. Implement error types
  - [x] 3.1 Implement `errors.go` with `ErrorCode` constants and `*Error` struct
    - Define all `ErrorCode` constants: `NetworkError`, `HTTPError`, `NotFound`, `NotSupported`, `InvalidInput`, `NoData`, `SerialiseError`, `ConfigError`
    - Implement `Error() string` and `Unwrap() error` methods
    - _Requirements: 13.1, 13.4_

  - [x] 3.2 Write property test for error unwrapping (Property 20)
    - Property 20: Error type is unwrappable via errors.As and errors.Is
    - Generate `*Error` values with arbitrary causes; assert `errors.As` and `errors.Is` traverse the chain correctly
    - _Validates: Requirements 13.4_

- [ ] 4. Implement HTTP client factory
  - [x] 4.1 Implement `internal/httpclient/httpclient.go`
    - `New(proxyURL string, timeout time.Duration) (*http.Client, error)`
    - Parse and validate proxy URL; configure `http.Transport` with proxy func
    - Set `Timeout` on the returned `*http.Client`
    - Return `*Error{Code: ConfigError}` for malformed proxy URL
    - _Requirements: 8.1, 8.3, 8.4_

  - [x] 4.2 Write unit tests for HTTP client factory
    - Test no-proxy path returns client with default transport
    - Test valid proxy URL configures transport correctly
    - Test malformed proxy URL returns `ConfigError`
    - _Requirements: 8.1, 8.3, 8.4_

- [ ] 5. Implement Config and Client constructor
  - [x] 5.1 Implement `Config` struct and `New()` in `vnstock.go`
    - `Config` fields: `Connector`, `ProxyURL`, `Timeout`, `MaxRetries`, `Logger`
    - `New(cfg Config) (*Client, error)`: apply env var defaults (`VNSTOCK_CONNECTOR`, `VNSTOCK_PROXY_URL`), apply field defaults (30s timeout, 3 retries, `slog.Default()`), validate all fields, construct named connector, return `*Error{Code: ConfigError}` for any invalid field
    - _Requirements: 12.1, 12.2, 12.3, 12.4, 7.3_

  - [x] 5.2 Write property test for invalid Config rejection (Property 13)
    - Property 13: Invalid Config fields are rejected at construction
    - Generate configs with negative timeout, malformed proxy URL, or empty connector name; assert `New` returns `*Error{Code: ConfigError}`
    - _Validates: Requirements 12.3_

  - [x] 5.3 Write property test for unrecognised connector rejection (Property 12)
    - Property 12: Unrecognised connector name is rejected at construction
    - Generate arbitrary strings not in `{"VCI","DNSE","FMP","Binance"}`; assert `New` returns `*Error{Code: ConfigError}`
    - _Validates: Requirements 7.3_

  - [x] 5.4 Write unit tests for Config defaults and env var overrides
    - Test zero-value Config applies documented defaults
    - Test `VNSTOCK_CONNECTOR` and `VNSTOCK_PROXY_URL` env vars are read when fields are empty
    - _Requirements: 12.2, 12.4_

- [x] 6. Define Connector interface
  - Implement `connector.go` with the full `Connector` interface
  - All eight methods: `QuoteHistory`, `RealTimeQuotes`, `Listing`, `IndexCurrent`, `IndexHistory`, `CompanyProfile`, `Officers`, `FinancialStatement`
  - Add `ErrNotSupported` sentinel and `MockConnector` stub (returns `ErrNotSupported` for all methods) for use in tests
  - _Requirements: 7.1, 7.4_

- [x] 7. Implement Client methods with request validation and delegation
  - [x] 7.1 Implement all `Client` methods in `vnstock.go`
    - Each method validates inputs, delegates to `c.connector`, wraps errors as `*Error`
    - `QuoteHistory`: validate `Start < End`, return `InvalidInput` error before any network call
    - `RealTimeQuotes`: fan-out concurrent requests via goroutines + `sync.WaitGroup`; aggregate results without shared mutable state
    - All methods wrap network errors as `NetworkError`, HTTP non-2xx as `HTTPError` with `StatusCode`
    - _Requirements: 1.1, 1.2, 1.3, 1.5, 2.1, 2.2, 2.3, 7.2, 11.1, 11.2, 13.2, 13.3_

  - [x] 7.2 Write property test for invalid date range rejection (Property 2)
    - Property 2: Invalid date range is rejected before network activity
    - Generate `QuoteHistoryRequest` where `Start > End`; assert `InvalidInput` error and no HTTP call made
    - _Validates: Requirements 1.3_

  - [x] 7.3 Write property test for concurrent safety (Property 4)
    - Property 4: Concurrent requests are free of data races
    - Spin up N goroutines calling mixed `Client` methods on the same instance; run under `-race`
    - _Validates: Requirements 2.3, 11.1, 11.2_

  - [x] 7.4 Write property test for named connector routing (Property 11)
    - Property 11: Named connector routes all requests through that connector
    - Use a recording `MockConnector`; assert every call is dispatched to it and no other connector receives calls
    - _Validates: Requirements 7.2_

  - [x] 7.5 Write unit tests for Client error wrapping
    - Test network failure → `NetworkError` with unwrappable cause
    - Test HTTP 404 → `HTTPError` with `StatusCode == 404`
    - Test `QuoteHistory` with `Start == End` → `InvalidInput`
    - _Requirements: 1.2, 1.3, 13.2, 13.3_

- [x] 8. Checkpoint — ensure all tests pass
  - Run `go test ./... -race`; ensure all tests pass, ask the user if questions arise.

- [x] 9. Implement VCI connector
  - [x] 9.1 Implement `connector/vci/vci.go`
    - Implement all `Connector` interface methods against VCI API endpoints
    - Use shared `*http.Client` from `internal/httpclient`
    - Return `ErrNotSupported` for any method VCI does not provide
    - Emit DEBUG log entry (URL, method, status, elapsed) for every request via injected `*slog.Logger`
    - _Requirements: 7.1, 7.4, 14.1, 14.4_

  - [x] 9.2 Write unit tests for VCI connector using `httptest.NewServer`
    - Test successful quote history response is parsed and ordered ascending by timestamp
    - Test HTTP 500 response returns `HTTPError` with correct status code
    - Test network failure returns `NetworkError`
    - _Requirements: 1.1, 1.2, 13.2, 13.3_

  - [x] 9.3 Write property test for quote history ordering (Property 1)
    - Property 1: Quote history is ordered ascending by timestamp
    - Generate arbitrary ordered/unordered quote slices from mock server; assert returned slice is sorted ascending
    - _Validates: Requirements 1.1_

- [x] 10. Implement DNSE connector
  - [x] 10.1 Implement `connector/dnse/dnse.go`
    - Implement all supported `Connector` methods against DNSE API endpoints
    - Return `ErrNotSupported` for unsupported methods
    - Emit DEBUG log entries for every request
    - _Requirements: 7.1, 7.4, 14.4_

  - [x] 10.2 Write unit tests for DNSE connector using `httptest.NewServer`
    - Test real-time quotes response covers all requested symbols
    - Test missing symbol returns `NotFound` error identifying the symbol
    - _Requirements: 2.1, 2.2_

  - [x] 10.3 Write property test for real-time quotes coverage (Property 3)
    - Property 3: Real-time quotes cover all requested symbols
    - Generate arbitrary symbol lists; assert returned slice has exactly one `Quote` per symbol with matching `Symbol` field
    - _Validates: Requirements 2.1_

- [x] 11. Implement FMP connector
  - [x] 11.1 Implement `connector/fmp/fmp.go`
    - Implement all supported `Connector` methods against FMP API endpoints
    - Return `ErrNotSupported` for unsupported methods
    - Emit DEBUG log entries for every request
    - _Requirements: 7.1, 7.4, 14.4_

  - [x] 11.2 Write unit tests for FMP connector using `httptest.NewServer`
    - Test company profile response populates all required fields
    - Test financial statement response is ordered descending by period
    - _Requirements: 5.1, 6.1_

  - [x] 11.3 Write property test for company profile fields (Property 8)
    - Property 8: Company profile contains all required fields
    - Generate arbitrary `CompanyProfile` values; assert all required fields are non-empty
    - _Validates: Requirements 5.1_

  - [x] 11.4 Write property test for financial statement ordering (Property 10)
    - Property 10: Financial statements are ordered descending by period
    - Generate arbitrary `FinancialPeriod` slices; assert returned slice is sorted descending by year then quarter
    - _Validates: Requirements 6.1_

- [x] 12. Implement Binance connector
  - [x] 12.1 Implement `connector/binance/binance.go`
    - Implement quote-related `Connector` methods against Binance API endpoints
    - Return `ErrNotSupported` for methods not applicable to crypto (e.g., `FinancialStatement`, `Officers`)
    - Emit DEBUG log entries for every request
    - _Requirements: 7.1, 7.4, 14.4_

  - [x] 12.2 Write unit tests for Binance connector using `httptest.NewServer`
    - Test `FinancialStatement` returns `ErrNotSupported`
    - Test successful quote history is ordered ascending by timestamp
    - _Requirements: 7.4, 1.1_

- [x] 13. Implement listing and index methods across connectors
  - [x] 13.1 Add `Listing` and `IndexHistory`/`IndexCurrent` implementations to connectors that support them
    - Ensure `Listing(ctx, exchange)` filters records by exchange before returning
    - Ensure `IndexHistory` returns records sorted ascending by timestamp
    - Return `NoData` error when listing is empty
    - Return `NotFound` error for unsupported index names with list of supported names
    - _Requirements: 3.1, 3.2, 3.3, 4.1, 4.2, 4.3_

  - [x] 13.2 Write property test for listing exchange filter (Property 5)
    - Property 5: Listing exchange filter returns only matching records
    - Generate arbitrary listing data with mixed exchanges; assert every returned record has `Exchange == requested`
    - _Validates: Requirements 3.2_

  - [x] 13.3 Write property test for listing record fields (Property 6)
    - Property 6: Listing records contain all required fields
    - Generate arbitrary `ListingRecord` slices; assert `Symbol`, `Exchange`, `CompanyName`, `Sector` are all non-empty
    - _Validates: Requirements 3.1_

  - [x] 13.4 Write property test for index history ordering (Property 7)
    - Property 7: Index history is ordered ascending by timestamp
    - Generate arbitrary `IndexRecord` slices; assert returned slice is sorted ascending by `Timestamp`
    - _Validates: Requirements 4.2_

  - [x] 13.5 Write property test for officer record fields (Property 9)
    - Property 9: Officer records contain all required fields
    - Generate arbitrary `Officer` slices; assert `Name`, `Title`, `AppointmentDate` are all non-empty
    - _Validates: Requirements 5.3_

- [x] 14. Implement exporters
  - [x] 14.1 Implement `exporter/csv.go`
    - `ExportCSV(w io.Writer, records any) error` using `encoding/csv`
    - Reflect over struct fields to write header row; write one data row per record
    - Return `*Error{Code: SerialiseError}` on write failure
    - _Requirements: 9.1, 9.5_

  - [x] 14.2 Implement `exporter/json.go`
    - `ExportJSON(w io.Writer, records any) error` using `encoding/json`
    - Write a JSON array; pretty-print with consistent indentation
    - Return `*Error{Code: SerialiseError}` on marshal/write failure
    - _Requirements: 9.2, 9.5, 10.2_

  - [x] 14.3 Implement `exporter/excel.go`
    - `ExportExcel(w io.Writer, records any) error` using `github.com/xuri/excelize/v2`
    - One worksheet per data type; header row + one row per record
    - Return `*Error{Code: SerialiseError}` on failure; return `*Error` identifying path on unwritable target
    - _Requirements: 9.3, 9.4, 9.5_

  - [x] 14.4 Write property test for CSV export structure (Property 14)
    - Property 14: CSV export produces valid structure for any data slice
    - Generate arbitrary non-empty slices of each exportable type; assert output parses as valid CSV with exactly one header row and one data row per record
    - _Validates: Requirements 9.1, 9.5_

  - [x] 14.5 Write property test for JSON export structure (Property 15)
    - Property 15: JSON export produces valid array for any data slice
    - Generate arbitrary non-empty slices; assert output parses as valid JSON array with element count matching input
    - _Validates: Requirements 9.2, 9.5_

  - [x] 14.6 Write unit tests for exporter error paths
    - Test unwritable `io.Writer` (e.g., always-error writer) returns `SerialiseError`
    - Test empty slice produces header-only CSV and empty JSON array
    - _Requirements: 9.4_

- [x] 15. Implement structured logging
  - [x] 15.1 Wire `*slog.Logger` through `Client` and all connectors
    - Each connector accepts a `*slog.Logger` at construction; logs URL, method, status, elapsed at DEBUG level after every request
    - `New()` passes `cfg.Logger` (or `slog.Default()`) to the constructed connector
    - _Requirements: 14.1, 14.2, 14.3, 14.4_

  - [x] 15.2 Write property test for debug log fields (Property 21)
    - Property 21: Debug log entries contain required fields for every request
    - Capture slog output via a `slog.Handler` backed by `bytes.Buffer`; for any connector request assert log record contains `url`, `method`, `status`, and `elapsed` attributes
    - _Validates: Requirements 14.4_

  - [x] 15.3 Write unit tests for logger injection
    - Test custom logger is used when provided in Config
    - Test default `slog.Default()` is used when `Config.Logger` is nil
    - _Requirements: 14.2, 14.3_

- [x] 16. Wire proxy support end-to-end
  - [x] 16.1 Integrate `internal/httpclient.New` into all connector constructors
    - Pass `cfg.ProxyURL` and `cfg.Timeout` when building each connector's `*http.Client`
    - _Requirements: 8.1, 8.2, 8.3, 8.4_

  - [x] 16.2 Write unit tests for proxy routing using `httptest.NewServer`
    - Stand up a local forwarding proxy server; configure `ProxyURL` to point at it; assert requests are routed through it
    - Test unreachable proxy returns `NetworkError` identifying the proxy as failure point
    - _Requirements: 8.1, 8.2_

  - [x] 16.3 Write property test for network error code (Property 18)
    - Property 18: Network errors carry NetworkError code
    - Simulate connection-refused / DNS failure; assert returned error is `*Error{Code: NetworkError}` with unwrappable cause
    - _Validates: Requirements 1.2, 13.2_

  - [x] 16.4 Write property test for HTTP error status code (Property 19)
    - Property 19: Non-2xx HTTP responses carry HTTP status code
    - Generate arbitrary non-2xx status codes via `httptest.NewServer`; assert `*Error{Code: HTTPError, StatusCode: <actual>}`
    - _Validates: Requirements 13.3_

- [x] 17. Final checkpoint — ensure all tests pass
  - Run `go test ./... -race -count=1`; ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for a faster MVP
- Each task references specific requirements for traceability
- Property tests use `pgregory.net/rapid`; tag each test with `// Feature: vnstock-go-rewrite, Property N: <text>`
- Checkpoints validate incremental correctness before moving to the next layer
