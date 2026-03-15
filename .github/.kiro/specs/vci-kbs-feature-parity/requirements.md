# Requirements Document

## Introduction

This feature implements missing VCI connector methods and adds a complete KBS connector to achieve feature parity with the Python vnstock library. The vnstock-go library currently has working VCI QuoteHistory and RealTimeQuotes implementations, but lacks listing, index, company profile, and financial statement capabilities. KBS is recommended as the primary connector due to its stability and comprehensive data coverage (30 columns for company data vs VCI's 10).

## Glossary

- **VCI_Connector**: The Vietcap Securities data source connector
- **KBS_Connector**: The KB Securities data source connector
- **Connector_Interface**: The Go interface defining standard data retrieval methods
- **Quote_Data**: Historical and real-time OHLCV (Open, High, Low, Close, Volume) price data
- **Listing_Data**: Information about all tradeable symbols on Vietnamese exchanges
- **Index_Data**: Market index values and historical performance
- **Company_Profile**: Descriptive information about listed companies including officers and shareholders
- **Financial_Statement**: Balance sheet, income statement, and cash flow data
- **HTTP_Client**: The shared HTTP client for making API requests
- **GraphQL_Endpoint**: VCI's GraphQL API endpoint for company and financial data
- **Exchange**: Vietnamese stock exchanges (HOSE, HNX, UPCOM)

## Requirements

### Requirement 1: KBS Connector Implementation

**User Story:** As a developer, I want a complete KBS connector, so that I can access stable and comprehensive Vietnamese stock market data.

#### Acceptance Criteria

1. THE KBS_Connector SHALL implement all methods defined in the Connector_Interface
2. WHEN the KBS_Connector is instantiated, THE KBS_Connector SHALL accept an HTTP_Client and logger
3. THE KBS_Connector SHALL provide QuoteHistory with support for daily, hourly, and minute intervals
4. THE KBS_Connector SHALL provide RealTimeQuotes for multiple symbols in a single request
5. THE KBS_Connector SHALL provide Listing data filterable by Exchange
6. THE KBS_Connector SHALL provide IndexCurrent for retrieving current market index values
7. THE KBS_Connector SHALL provide IndexHistory for retrieving historical market index data
8. THE KBS_Connector SHALL provide CompanyProfile with at least 30 data columns
9. THE KBS_Connector SHALL provide Officers data including executive names and positions
10. THE KBS_Connector SHALL provide FinancialStatement data for balance sheet, income statement, and cash flow
11. WHEN an API request fails, THE KBS_Connector SHALL return a descriptive error with appropriate error code
12. THE KBS_Connector SHALL use the shared HTTP_Client from internal/httpclient package
13. THE KBS_Connector SHALL log all API requests with method, URL, status code, and elapsed time

### Requirement 2: VCI Listing Implementation

**User Story:** As a developer, I want to retrieve stock listings from VCI, so that I can get a complete list of tradeable symbols.

#### Acceptance Criteria

1. THE VCI_Connector SHALL implement the Listing method from Connector_Interface
2. WHEN an empty exchange parameter is provided, THE VCI_Connector SHALL return all symbols across all exchanges
3. WHEN a specific Exchange is provided, THE VCI_Connector SHALL return only symbols from that Exchange
4. THE VCI_Connector SHALL return Listing_Data including symbol, company name, exchange, and industry classification
5. WHEN the VCI API returns an error, THE VCI_Connector SHALL return an error with code API_ERROR

### Requirement 3: VCI Index Data Implementation

**User Story:** As a developer, I want to retrieve market index data from VCI, so that I can track market performance.

#### Acceptance Criteria

1. THE VCI_Connector SHALL implement the IndexCurrent method from Connector_Interface
2. THE VCI_Connector SHALL implement the IndexHistory method from Connector_Interface
3. WHEN IndexCurrent is called with a valid index name, THE VCI_Connector SHALL return current Index_Data
4. WHEN IndexHistory is called with valid parameters, THE VCI_Connector SHALL return historical Index_Data
5. THE VCI_Connector SHALL support VN-Index, HNX-Index, and UPCOM-Index
6. WHEN an invalid index name is provided, THE VCI_Connector SHALL return an error with code INVALID_INPUT

### Requirement 4: VCI Company Profile Implementation

**User Story:** As a developer, I want to retrieve company information from VCI, so that I can analyze company fundamentals.

#### Acceptance Criteria

1. THE VCI_Connector SHALL implement the CompanyProfile method from Connector_Interface
2. THE VCI_Connector SHALL implement the Officers method from Connector_Interface
3. WHEN CompanyProfile is called with a valid symbol, THE VCI_Connector SHALL return Company_Profile data
4. WHEN Officers is called with a valid symbol, THE VCI_Connector SHALL return a list of company officers
5. THE VCI_Connector SHALL use the GraphQL_Endpoint at https://trading.vietcap.com.vn/data-mt/graphql
6. THE VCI_Connector SHALL include company overview, shareholders, and ownership structure in Company_Profile
7. WHEN an invalid symbol is provided, THE VCI_Connector SHALL return an error with code INVALID_INPUT
8. WHEN the GraphQL_Endpoint returns an error, THE VCI_Connector SHALL return an error with code API_ERROR

### Requirement 5: VCI Financial Statement Implementation

**User Story:** As a developer, I want to retrieve financial statements from VCI, so that I can perform fundamental analysis.

#### Acceptance Criteria

1. THE VCI_Connector SHALL implement the FinancialStatement method from Connector_Interface
2. WHEN FinancialStatement is called with statement type "balance_sheet", THE VCI_Connector SHALL return balance sheet data
3. WHEN FinancialStatement is called with statement type "income_statement", THE VCI_Connector SHALL return income statement data
4. WHEN FinancialStatement is called with statement type "cash_flow", THE VCI_Connector SHALL return cash flow data
5. THE VCI_Connector SHALL support both annual and quarterly period types
6. THE VCI_Connector SHALL use the GraphQL_Endpoint for financial data retrieval
7. WHEN an invalid statement type is provided, THE VCI_Connector SHALL return an error with code INVALID_INPUT
8. THE VCI_Connector SHALL return financial data sorted by period in descending order

### Requirement 6: Connector Registration

**User Story:** As a developer, I want KBS to be available as a connector option, so that I can use it through the vnstock client.

#### Acceptance Criteria

1. THE KBS_Connector SHALL be registered in the connector registry
2. WHEN a client is created with connector name "kbs", THE Client SHALL use the KBS_Connector
3. WHEN a client is created with connector name "vci", THE Client SHALL use the VCI_Connector with all implemented methods
4. THE Client SHALL return an error with code INVALID_INPUT when an unsupported connector name is provided

### Requirement 7: Error Handling and Validation

**User Story:** As a developer, I want clear error messages, so that I can debug issues quickly.

#### Acceptance Criteria

1. WHEN a connector method receives invalid date ranges, THE Connector SHALL return an error with code INVALID_INPUT
2. WHEN a connector method receives empty required parameters, THE Connector SHALL return an error with code INVALID_INPUT
3. WHEN an API request times out, THE Connector SHALL return an error with code NETWORK_ERROR
4. WHEN an API returns HTTP status 4xx, THE Connector SHALL return an error with code API_ERROR
5. WHEN an API returns HTTP status 5xx, THE Connector SHALL return an error with code API_ERROR
6. THE Connector SHALL include the original error message in the error response
7. THE Connector SHALL validate symbol format before making API requests

### Requirement 8: Testing and Quality Assurance

**User Story:** As a developer, I want comprehensive tests, so that I can trust the connector implementations.

#### Acceptance Criteria

1. THE KBS_Connector SHALL have unit tests for all Connector_Interface methods
2. THE VCI_Connector SHALL have unit tests for all newly implemented methods
3. WHEN unit tests are executed, THE tests SHALL use mocked HTTP responses
4. THE tests SHALL verify correct API endpoint URLs are called
5. THE tests SHALL verify correct request payloads are sent
6. THE tests SHALL verify response parsing produces correct data structures
7. THE tests SHALL verify error conditions return appropriate error codes
8. THE tests SHALL achieve at least 80% code coverage for connector packages
9. THE KBS_Connector SHALL have integration tests that call real KBS APIs
10. THE VCI_Connector SHALL have integration tests for newly implemented methods

### Requirement 9: Documentation and Examples

**User Story:** As a developer, I want clear documentation, so that I can use the connectors effectively.

#### Acceptance Criteria

1. THE KBS_Connector SHALL have package-level documentation describing its purpose
2. THE KBS_Connector SHALL have method-level documentation for all public methods
3. THE project SHALL include example code demonstrating KBS_Connector usage
4. THE project SHALL include example code demonstrating VCI_Connector usage for new methods
5. THE FEATURE_COVERAGE.md file SHALL be updated to reflect implemented features
6. THE API_ENDPOINTS.md file SHALL document KBS API endpoints and request/response formats

### Requirement 10: Data Model Compatibility

**User Story:** As a developer, I want consistent data models, so that I can switch between connectors easily.

#### Acceptance Criteria

1. THE KBS_Connector SHALL return Quote data using the existing Quote model
2. THE KBS_Connector SHALL return Listing_Data using the existing ListingRecord model
3. THE KBS_Connector SHALL return Index_Data using the existing IndexRecord model
4. THE KBS_Connector SHALL return Company_Profile using the existing CompanyProfile model
5. THE KBS_Connector SHALL return Officers using the existing Officer model
6. THE KBS_Connector SHALL return Financial_Statement using the existing FinancialPeriod model
7. WHEN data models need additional fields, THE models SHALL be extended without breaking existing code
8. THE models SHALL use appropriate Go types for financial data (float64 for prices, int64 for volumes)

### Requirement 11: Performance and Reliability

**User Story:** As a developer, I want reliable and performant connectors, so that my applications run smoothly.

#### Acceptance Criteria

1. WHEN making API requests, THE Connector SHALL set appropriate timeout values
2. WHEN making concurrent requests, THE Connector SHALL be thread-safe
3. THE Connector SHALL reuse HTTP connections through the shared HTTP_Client
4. WHEN API responses are large, THE Connector SHALL stream and parse data efficiently
5. THE Connector SHALL log request timing for performance monitoring
6. WHEN rate limits are encountered, THE Connector SHALL return an error with code RATE_LIMITED

### Requirement 12: API Endpoint Research and Implementation

**User Story:** As a developer, I want accurate API implementations, so that data retrieval works correctly.

#### Acceptance Criteria

1. THE implementation SHALL research KBS API endpoints from Python vnstock source code
2. THE implementation SHALL document discovered KBS API endpoints in API_ENDPOINTS.md
3. THE implementation SHALL verify API endpoint behavior through manual testing
4. THE implementation SHALL document request payload formats for all endpoints
5. THE implementation SHALL document response formats for all endpoints
6. THE implementation SHALL handle API versioning if multiple versions exist
