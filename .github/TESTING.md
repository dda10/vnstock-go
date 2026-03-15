# GitHub Actions Testing Integration

Guide for vnstock-go's automated testing and CI/CD using GitHub Actions.

## Workflows

### 1. Test Suite (`test.yml`)

**Trigger**: Push/PR to main/develop

- Runs `go test -race` across Ubuntu, macOS, Windows
- Go versions: 1.22, 1.23, 1.24
- Generates coverage on ubuntu/1.24 and uploads to Codecov

### 2. Coverage Report (`coverage-report.yml`)

**Trigger**: Push/PR to main/develop

- Generates coverage with `go test -coverprofile`
- Displays summary via `go tool cover -func`
- Uploads to Codecov

### 3. Code Quality (`code-quality.yml`)

**Trigger**: Push/PR to main/develop

- `go vet` for static analysis
- `gofmt` formatting check
- `golangci-lint` for comprehensive linting
- `govulncheck` for dependency vulnerabilities

### 4. Performance Testing (`performance.yml`)

**Trigger**: Push/PR to main/develop, daily at 3 AM UTC

- Runs `go test -bench` with memory stats
- Stores benchmark results as artifacts

### 5. Verify Build (`verify-api-key.yml`)

**Trigger**: Manual, weekly, or on workflow file changes

- Verifies the project builds successfully
- Runs a basic smoke test

## Running Tests Locally

```bash
cd vnstock-go

# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run specific package
go test ./connector/vci/...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Run benchmarks
go test -bench=. -benchmem ./...

# Verbose output
go test -v ./...
```

## PR Checks

Pull requests automatically run:
1. ✅ Tests pass across all OS/Go version combinations
2. ✅ Code passes `go vet`, `gofmt`, and `golangci-lint`
3. ✅ No known vulnerabilities in dependencies
4. ✅ Build succeeds
