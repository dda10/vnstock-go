# Tech Stack

## Go Library (vnstock-go)

### Language & Runtime
- Go 1.22+
- Standard library: `net/http`, `encoding/json`, `log/slog`
- External: `github.com/xuri/excelize/v2` (Excel export)

### Module Path
- `github.com/dda10/vnstock-go`

### Testing
- Standard `go test`
- Property-based testing with `pgregory.net/rapid`
- Race detection with `-race` flag

### Common Commands

```bash
cd vnstock-go

# Run tests
go test ./...

# Run with race detector
go test -race ./...

# Run specific package tests
go test ./connector/vci/...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Run benchmarks
go test -bench=. -benchmem ./...

# Build
go build ./...

# Format code
go fmt ./...

# Vet for issues
go vet ./...
```

## Data Sources

| Source  | Type       | Coverage                          |
|---------|------------|-----------------------------------|
| VCI     | Vietnamese | Quotes, listings, financials      |
| KBS     | Vietnamese | Quotes, financials, trading       |
| DNSE    | Vietnamese | Quotes, company data              |
| FMP     | International | Global stocks, forex           |
| Binance | Crypto     | Cryptocurrency prices             |
| Gold    | Precious metals | Domestic & international gold |

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `VNSTOCK_CONNECTOR` | Default connector name |
| `VNSTOCK_PROXY_URL` | HTTP proxy URL |
