# Project Structure

## Repository Layout

```
vnstock-go/
├── .github/
│   ├── workflows/          # CI/CD pipelines
│   └── TESTING.md          # Testing guide
├── vnstock-go/             # Go module root
│   ├── go.mod
│   ├── go.sum
│   ├── vnstock.go          # Client, Config, New()
│   ├── connector.go        # Connector interface
│   ├── errors.go           # Error types and codes
│   ├── models.go           # Data models (Quote, Listing, etc.)
│   ├── registry.go         # Connector registry
│   ├── connector/          # Data source implementations
│   │   ├── vci/            # VCI connector
│   │   ├── kbs/            # KBS connector
│   │   ├── dnse/           # DNSE connector
│   │   ├── fmp/            # FMP connector
│   │   ├── gold/           # Gold price connector
│   │   └── binance/        # Binance connector
│   ├── exporter/           # CSV, JSON, Excel exporters
│   ├── examples/           # Usage examples
│   │   ├── basic/
│   │   ├── library-test/
│   │   └── api-server/
│   ├── all/                # Convenience import for all connectors
│   └── internal/
│       └── httpclient/     # Shared HTTP client
├── .gitignore
├── CHANGELOG.md
└── README.md
```

## Go Package Conventions

- `vnstock-go/`: Root package exports public API (`Client`, `Config`, models)
- `connector/`: Each data source in its own subpackage implementing `Connector` interface
- `internal/`: Private packages not exported to callers
- `exporter/`: Export utilities for CSV, JSON, Excel
- `all/`: Side-effect import to register all connectors

## Key Files

| File | Purpose |
|------|---------|
| `vnstock.go` | Client entry point, `New()` constructor |
| `connector.go` | `Connector` interface definition |
| `models.go` | Shared data models |
| `errors.go` | Error types and codes |
| `registry.go` | Connector registration |

## Naming Conventions

- PascalCase for exported symbols, camelCase for unexported
- Test files: `*_test.go`
- Connectors: Named after data source (vci, kbs, dnse, fmp, binance, gold)
