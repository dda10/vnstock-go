# Implementation Plan: VCI-KBS Feature Parity

## Overview

This implementation adds missing VCI connector methods and creates a complete KBS connector to achieve feature parity with the Python vnstock library. The work is divided into three main areas: KBS connector implementation (primary data source), VCI connector completion, and comprehensive testing.

## Tasks

- [x] 1. Research and document KBS API endpoints
  - Analyze Python vnstock KBS implementation for endpoint URLs
  - Test endpoints manually to verify request/response formats
  - Document all endpoints in `vnstock-go/API_ENDPOINTS.md`
  - Document authentication requirements (API keys, tokens)
  - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5_

- [x] 2. Create KBS connector package structure
  - [x] 2.1 Create `vnstock-go/connector/kbs/` directory
    - Create `kbs.go` with Connector struct and New() function
    - Implement connector registration in init() function
    - Add shared HTTP client and logger fields
    - _Requirements: 1.1, 1.2, 1.12_
  
  - [x] 2.2 Implement KBS helper functions
    - Implement doRequest() for HTTP requests with logging
    - Implement logRequest() for structured logging
    - Implement validateSymbol() for symbol format validation
    - Implement validateDateRange() for date validation
    - _Requirements: 1.11, 1.13, 7.1, 7.2, 7.7_

- [] 3. Implement KBS quote methods
  - [x] 3.1 Implement KBS QuoteHistory
    - Build request payload with symbol, start, end, interval
    - Parse JSON response to Quote slice
    - Map interval values correctly
    - Handle empty responses with NoData error
    - _Requirements: 1.3_
  
  - [ ]* 3.2 Write property test for KBS QuoteHistory
    - **Property 1: Quote Interval Preservation**
    - **Validates: Requirements 1.3**
  
  - [x] 3.3 Implement KBS RealTimeQuotes
    - Build request with multiple symbols
    - Parse JSON response to Quote slice
    - Validate all returned symbols match requested symbols
    - _Requirements: 1.4_
  
  - [ ]* 3.4 Write property test for KBS RealTimeQuotes
    - **Property 2: Real-Time Quote Symbol Matching**
    - **Validates: Requirements 1.4**

- [ ] 4. Implement KBS listing and index methods
  - [x] 4.1 Implement KBS Listing
    - Build request with optional exchange filter
    - Parse JSON response to ListingRecord slice
    - Filter by exchange if parameter provided
    - Validate required fields are populated
    - _Requirements: 1.5, 10.2_
  
  - [ ]* 4.2 Write property tests for KBS Listing
    - **Property 3: Listing Exchange Filtering**
    - **Property 11: Listing Required Fields**
    - **Validates: Requirements 1.5, 10.2_
  
  - [x] 4.3 Implement KBS IndexCurrent
    - Build request with index name
    - Parse JSON response to IndexRecord
    - Validate index name before request
    - _Requirements: 1.6_
  
  - [ ]* 4.4 Write property test for KBS IndexCurrent
    - **Property 4: Index Data Retrieval**
    - **Validates: Requirements 1.6**
  
  - [x] 4.5 Implement KBS IndexHistory
    - Build request with index name, start, end
    - Parse JSON response to IndexRecord slice
    - Validate date range
    - _Requirements: 1.7_
  
  - [ ]* 4.6 Write property test for KBS IndexHistory
    - **Property 5: Index History Date Range**
    - **Validates: Requirements 1.7**

- [ ] 5. Implement KBS company data methods
  - [x] 5.1 Implement KBS CompanyProfile
    - Build request with symbol
    - Parse JSON response to CompanyProfile
    - Ensure at least 30 data columns are populated
    - Validate required fields (Symbol, Name, Exchange, Sector)
    - _Requirements: 1.8, 10.4_
  
  - [ ]* 5.2 Write property test for KBS CompanyProfile
    - **Property 6: Company Profile Field Population**
    - **Validates: Requirements 1.8**
  
  - [x] 5.3 Implement KBS Officers
    - Build request with symbol
    - Parse JSON response to Officer slice
    - Validate Name and Title fields are populated
    - _Requirements: 1.9, 10.5_
  
  - [ ]* 5.4 Write property test for KBS Officers
    - **Property 7: Officers List Non-Empty**
    - **Validates: Requirements 1.9**

- [ ] 6. Implement KBS financial statement method
  - [x] 6.1 Implement KBS FinancialStatement
    - Build request with symbol, statement type, period
    - Validate statement type (income, balance, cashflow)
    - Parse JSON response to FinancialPeriod slice
    - Sort results by (Year, Quarter) descending
    - Map financial fields to Fields map
    - _Requirements: 1.10, 10.6_
  
  - [ ]* 6.2 Write property tests for KBS FinancialStatement
    - **Property 8: Financial Statement Type Handling**
    - **Property 9: Financial Period Support**
    - **Property 10: Financial Data Ordering**
    - **Validates: Requirements 1.10**

- [x] 7. Checkpoint - KBS connector complete
  - Ensure all KBS tests pass, ask the user if questions arise.

- [ ] 8. Implement VCI Listing method
  - [x] 8.1 Research VCI listing endpoint
    - Determine if REST or GraphQL endpoint
    - Document request/response format
    - Test manually with sample requests
    - _Requirements: 2.1, 12.3_
  
  - [x] 8.2 Implement VCI Listing
    - Replace NotSupported stub in vci.go
    - Build request with optional exchange filter
    - Parse response to ListingRecord slice
    - Handle API errors appropriately
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_
  
  - [ ]* 8.3 Write property test for VCI Listing
    - **Property 3: Listing Exchange Filtering**
    - **Property 11: Listing Required Fields**
    - **Validates: Requirements 2.3, 2.4**

- [ ] 9. Implement VCI index methods
  - [x] 9.1 Research VCI index endpoints
    - Determine GraphQL queries or REST endpoints
    - Document request/response format
    - Test with VN-Index, HNX-Index, UPCOM-Index
    - _Requirements: 3.1, 3.2, 12.3_
  
  - [x] 9.2 Implement VCI IndexCurrent
    - Replace NotSupported stub in vci.go
    - Build GraphQL query or REST request
    - Map index names (VN-Index, HNX-Index, UPCOM-Index)
    - Parse response to IndexRecord
    - Validate index name before request
    - _Requirements: 3.1, 3.3, 3.6_
  
  - [ ]* 9.3 Write property test for VCI IndexCurrent
    - **Property 4: Index Data Retrieval**
    - **Property 18: Invalid Index Name Rejection**
    - **Validates: Requirements 3.3, 3.6**
  
  - [x] 9.4 Implement VCI IndexHistory
    - Replace NotSupported stub in vci.go
    - Build GraphQL query or REST request
    - Parse response to IndexRecord slice
    - Validate date range
    - _Requirements: 3.2, 3.4_
  
  - [ ]* 9.5 Write property test for VCI IndexHistory
    - **Property 5: Index History Date Range**
    - **Validates: Requirements 3.4**

- [x] 10. Implement VCI company profile methods
  - [x] 10.1 Create VCI GraphQL helper function
    - Implement doGraphQLRequest() for GraphQL queries
    - Handle GraphQL-specific error responses
    - Add to vci.go as private helper
    - _Requirements: 4.5_
  
  - [x] 10.2 Implement VCI CompanyProfile
    - Replace NotSupported stub in vci.go
    - Build GraphQL query for company data
    - Parse response including overview, shareholders, ownership
    - Map to CompanyProfile struct
    - _Requirements: 4.1, 4.3, 4.6, 4.7, 4.8_
  
  - [ ]* 10.3 Write property test for VCI CompanyProfile
    - **Property 6: Company Profile Field Population**
    - **Validates: Requirements 4.3**
  
  - [x] 10.4 Implement VCI Officers
    - Replace NotSupported stub in vci.go
    - Build GraphQL query for officers data
    - Parse response to Officer slice
    - _Requirements: 4.2, 4.4, 4.7, 4.8_
  
  - [ ]* 10.5 Write property test for VCI Officers
    - **Property 7: Officers List Non-Empty**
    - **Validates: Requirements 4.4**

- [x] 11. Implement VCI financial statement method
  - [x] 11.1 Implement VCI FinancialStatement
    - Replace NotSupported stub in vci.go
    - Build GraphQL query with statement type and period
    - Validate statement type (income, balance, cashflow)
    - Parse response to FinancialPeriod slice
    - Sort results by period descending
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7, 5.8_
  
  - [ ]* 11.2 Write property tests for VCI FinancialStatement
    - **Property 8: Financial Statement Type Handling**
    - **Property 9: Financial Period Support**
    - **Property 10: Financial Data Ordering**
    - **Property 19: Invalid Statement Type Rejection**
    - **Validates: Requirements 5.2, 5.3, 5.4, 5.5, 5.7, 5.8**

- [x] 12. Checkpoint - VCI connector complete
  - Ensure all VCI tests pass, ask the user if questions arise.

- [ ] 13. Update connector registration
  - [ ] 13.1 Register KBS connector
    - Verify KBS init() function registers connector
    - Update validConnectors map in vnstock.go if needed
    - Test client creation with "kbs" connector name
    - _Requirements: 6.1, 6.2_
  
  - [ ] 13.2 Verify VCI connector registration
    - Test client creation with "vci" connector name
    - Verify all new methods are accessible
    - _Requirements: 6.3_
  
  - [ ]* 13.3 Write property tests for connector registration
    - **Property 20: Connector Registration**
    - **Property 21: Invalid Connector Name Rejection**
    - **Validates: Requirements 6.2, 6.3, 6.4**

- [ ] 14. Implement comprehensive error handling
  - [ ] 14.1 Add error handling tests for KBS
    - Test invalid date ranges return InvalidInput
    - Test empty parameters return InvalidInput
    - Test HTTP 4xx/5xx return HTTPError
    - Test network failures return NetworkError
    - Test symbol format validation
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7_
  
  - [ ] 14.2 Add error handling tests for VCI
    - Test invalid index names return InvalidInput
    - Test invalid statement types return InvalidInput
    - Test GraphQL errors return API_ERROR
    - Test error message preservation
    - _Requirements: 3.6, 4.7, 4.8, 5.7, 7.6_
  
  - [ ]* 14.3 Write property tests for error handling
    - **Property 12: Invalid Date Range Rejection**
    - **Property 13: Empty Parameter Rejection**
    - **Property 14: HTTP Error Code Mapping**
    - **Property 15: Network Error Handling**
    - **Property 16: Error Message Preservation**
    - **Property 17: Symbol Format Validation**
    - **Property 18: Invalid Index Name Rejection**
    - **Property 19: Invalid Statement Type Rejection**
    - **Validates: Requirements 7.1-7.7, 3.6, 5.7**

- [ ] 15. Write unit tests for KBS connector
  - [ ] 15.1 Create KBS unit test file
    - Create `vnstock-go/connector/kbs/kbs_test.go`
    - Set up httptest.Server for mocking
    - Create mock response fixtures
    - _Requirements: 8.1, 8.3_
  
  - [ ] 15.2 Write KBS unit tests
    - Test QuoteHistory with mocked responses
    - Test RealTimeQuotes with mocked responses
    - Test Listing with mocked responses
    - Test IndexCurrent with mocked responses
    - Test IndexHistory with mocked responses
    - Test CompanyProfile with mocked responses
    - Test Officers with mocked responses
    - Test FinancialStatement with mocked responses
    - Verify correct endpoint URLs are called
    - Verify correct request payloads are sent
    - Verify response parsing produces correct structures
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7_

- [ ] 16. Write unit tests for VCI new methods
  - [ ] 16.1 Extend VCI unit test file
    - Add tests to `vnstock-go/connector/vci/vci_test.go`
    - Create mock GraphQL responses
    - _Requirements: 8.2, 8.3_
  
  - [ ] 16.2 Write VCI unit tests for new methods
    - Test Listing with mocked responses
    - Test IndexCurrent with mocked responses
    - Test IndexHistory with mocked responses
    - Test CompanyProfile with mocked GraphQL
    - Test Officers with mocked GraphQL
    - Test FinancialStatement with mocked GraphQL
    - Verify GraphQL query structure
    - _Requirements: 8.2, 8.3, 8.4, 8.5, 8.6_

- [ ] 17. Write integration tests
  - [ ] 17.1 Create KBS integration test file
    - Create `vnstock-go/connector/kbs/kbs_integration_test.go`
    - Add `//go:build integration` tag
    - Use real Vietnamese symbols (VNM, VCB, HPG)
    - _Requirements: 8.9_
  
  - [ ] 17.2 Write KBS integration tests
    - Test all Connector methods with real KBS API
    - Verify data structure and field population
    - Test with multiple exchanges and symbols
    - _Requirements: 8.9_
  
  - [ ] 17.3 Create VCI integration test file
    - Create `vnstock-go/connector/vci/vci_integration_test.go`
    - Add `//go:build integration` tag
    - _Requirements: 8.10_
  
  - [ ] 17.4 Write VCI integration tests for new methods
    - Test Listing with real VCI API
    - Test IndexCurrent with real VCI API
    - Test IndexHistory with real VCI API
    - Test CompanyProfile with real VCI GraphQL
    - Test Officers with real VCI GraphQL
    - Test FinancialStatement with real VCI GraphQL
    - _Requirements: 8.10_

- [ ] 18. Add concurrency safety tests
  - [ ]* 18.1 Write property test for concurrent requests
    - **Property 22: Concurrent Request Safety**
    - Test KBS connector with concurrent goroutines
    - Test VCI connector with concurrent goroutines
    - Run with `-race` flag to detect data races
    - **Validates: Requirements 11.2**

- [ ] 19. Update documentation
  - [ ] 19.1 Add KBS package documentation
    - Add package-level doc comment to kbs.go
    - Add method-level doc comments for all public methods
    - _Requirements: 9.1, 9.2_
  
  - [ ] 19.2 Create KBS usage examples
    - Create example in `vnstock-go/examples/kbs-demo/`
    - Demonstrate QuoteHistory, Listing, CompanyProfile
    - Add to examples README
    - _Requirements: 9.3_
  
  - [ ] 19.3 Create VCI usage examples for new methods
    - Extend existing VCI examples
    - Demonstrate Listing, IndexCurrent, CompanyProfile
    - _Requirements: 9.4_
  
  - [ ] 19.4 Update FEATURE_COVERAGE.md
    - Mark KBS connector as fully implemented
    - Mark VCI connector methods as implemented
    - Update feature parity status
    - _Requirements: 9.5_
  
  - [ ] 19.5 Update API_ENDPOINTS.md
    - Document all KBS endpoints with request/response formats
    - Document VCI GraphQL queries
    - Add authentication details
    - _Requirements: 9.6, 12.4, 12.5_

- [ ] 20. Final checkpoint - All tests pass
  - Run all unit tests: `go test ./...`
  - Run with race detector: `go test -race ./...`
  - Verify code coverage meets 80% target
  - Run integration tests manually: `go test -tags=integration ./...`
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional property-based tests that can be skipped for faster MVP
- Each task references specific requirements for traceability
- KBS connector is prioritized as the primary data source due to stability
- VCI connector completion enables fallback and comparison capabilities
- Property tests use `pgregory.net/rapid` with minimum 100 iterations
- Integration tests require manual execution and are excluded from CI/CD
- All code examples use Go 1.22+ with standard library and existing dependencies
