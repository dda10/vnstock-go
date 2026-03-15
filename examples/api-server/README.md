# Vnstock Go API Server

HTTP REST API wrapper for testing vnstock-go library with Postman or any HTTP client.

## Quick Start

### 1. Start the API Server

```bash
cd vnstock-go/examples/api-server
go run main.go
```

The server will start on `http://localhost:8080`

### 2. Test with Postman

#### Option A: Import Collection File
1. Open Postman
2. Click "Import" button
3. Select `vnstock-api.postman_collection.json`
4. All endpoints will be ready to test

#### Option B: Manual Testing
Use these endpoints directly in Postman:

**Health Check**
```
GET http://localhost:8080/health
```

**Get Historical Quotes**
```
GET http://localhost:8080/api/quote/history?symbol=VNM&days=30
```

**Get Real-Time Quotes**
```
GET http://localhost:8080/api/quote/realtime?symbols=VNM,HPG,VIC
```

**Get Market Listing**
```
GET http://localhost:8080/api/listing?exchange=HOSE
```

**Get Company Profile**
```
GET http://localhost:8080/api/company/profile?symbol=VNM
```

**Get Current Index**
```
GET http://localhost:8080/api/index/current?name=VNINDEX
```

**Get Index History**
```
GET http://localhost:8080/api/index/history?name=VNINDEX&days=30
```

## API Endpoints

| Endpoint | Method | Parameters | Description |
|----------|--------|------------|-------------|
| `/health` | GET | - | Health check |
| `/api/quote/history` | GET | `symbol`, `days` | Historical OHLCV data |
| `/api/quote/realtime` | GET | `symbols` (comma-separated) | Real-time quotes |
| `/api/listing` | GET | `exchange` (optional) | Market listings |
| `/api/company/profile` | GET | `symbol` | Company information |
| `/api/index/current` | GET | `name` | Current index value |
| `/api/index/history` | GET | `name`, `days` | Historical index data |

## Example Requests

### Using curl

```bash
# Health check
curl http://localhost:8080/health

# Get VNM stock history (last 30 days)
curl "http://localhost:8080/api/quote/history?symbol=VNM&days=30"

# Get real-time quotes for multiple symbols
curl "http://localhost:8080/api/quote/realtime?symbols=VNM,HPG,VIC"

# Get HOSE exchange listing
curl "http://localhost:8080/api/listing?exchange=HOSE"

# Get company profile
curl "http://localhost:8080/api/company/profile?symbol=VNM"

# Get current VNINDEX value
curl "http://localhost:8080/api/index/current?name=VNINDEX"

# Get VNINDEX history
curl "http://localhost:8080/api/index/history?name=VNINDEX&days=30"
```

## Response Format

All responses are in JSON format.

### Success Response Example (Quote History)
```json
[
  {
    "symbol": "VNM",
    "time": "2024-01-15T00:00:00Z",
    "open": 85000,
    "high": 86000,
    "low": 84500,
    "close": 85500,
    "volume": 1234567
  }
]
```

### Error Response Example
```json
{
  "error": "symbol parameter is required"
}
```

## Common Stock Symbols

- VNM: Vinamilk
- HPG: Hoa Phat Group
- VIC: Vingroup
- VCB: Vietcombank
- FPT: FPT Corporation
- MSN: Masan Group

## Common Exchanges

- HOSE: Ho Chi Minh Stock Exchange
- HNX: Hanoi Stock Exchange
- UPCOM: Unlisted Public Company Market

## Common Indices

- VNINDEX: VN-Index (HOSE)
- HNX-INDEX: HNX-Index
- UPCOM-INDEX: UPCOM-Index
