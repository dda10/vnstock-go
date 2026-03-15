# Real API Endpoints from vnstock Python Library

## VCI (Vietcap Securities)

**Base URL**: `https://trading.vietcap.com.vn/api/`

### Endpoints:

1. **Quote History (OHLC)**
   - URL: `POST https://trading.vietcap.com.vn/api/chart/OHLCChart/gap-chart`
   - Payload:
     ```json
     {
       "timeFrame": "ONE_DAY",  // ONE_MINUTE, ONE_HOUR, ONE_DAY
       "symbols": ["VNM"],
       "to": 1234567890,  // Unix timestamp
       "countBack": 100
     }
     ```

2. **Intraday (Real-time ticks)**
   - URL: `POST https://trading.vietcap.com.vn/api/market-watch/LEData/getAll`
   - Payload:
     ```json
     {
       "symbol": "VNM",
       "limit": 1000,
       "truncTime": 1234567890  // Unix timestamp
     }
     ```

3. **GraphQL Endpoint** (for company data, financials, etc.)
   - URL: `POST https://trading.vietcap.com.vn/data-mt/graphql`
   - This is a GraphQL endpoint - requires specific queries

### Interval Mapping:
- `1m`, `5m`, `15m`, `30m` → `ONE_MINUTE`
- `1H` → `ONE_HOUR`
- `1D`, `1W`, `1M` → `ONE_DAY`

### Response Format (OHLC):
```json
[
  {
    "t": 1234567890,  // timestamp
    "o": 85000,       // open
    "h": 86000,       // high
    "l": 84500,       // low
    "c": 85500,       // close
    "v": 1234567      // volume
  }
]
```

## DNSE

**Note**: DNSE API endpoints are not in the vnstock Python repository. The Python library uses a different approach or the endpoints are proprietary.

## KBS (KB Securities)

**Note**: KBS is mentioned in the Python library but specific API endpoints need further research.

## TCBS (Techcombank Securities)

**Note**: TCBS is mentioned in the Python library but specific API endpoints need further research.

## Implementation Notes

1. **Headers**: The Python library uses custom headers with user agents
2. **Proxy Support**: The Python library has built-in proxy support
3. **Rate Limiting**: No explicit rate limiting mentioned, but should be implemented
4. **Authentication**: No API key required for basic endpoints
5. **Error Handling**: APIs return JSON with error messages

## Next Steps

To implement real data fetching:

1. Update VCI connector with real endpoints
2. Research DNSE, KBS, TCBS API endpoints
3. Implement proper request/response handling
4. Add rate limiting
5. Add comprehensive error handling

## KBS (KB Securities)

**Base URLs**:
- IIS Server: `https://kbbuddywts.kbsec.com.vn/iis-server/investment`
- SAS Server: `https://kbbuddywts.kbsec.com.vn/sas`

### Authentication
- No API key required for public endpoints
- Uses standard HTTP headers with User-Agent

### Quote Endpoints

#### 1. Historical Quotes (OHLCV)
- **URL**: `GET https://kbbuddywts.kbsec.com.vn/iis-server/investment/stocks/{symbol}/data_{interval}`
- **URL (Index)**: `GET https://kbbuddywts.kbsec.com.vn/iis-server/investment/index/{symbol}/data_{interval}`
- **Intervals**: 
  - `1P` (1 minute)
  - `5P` (5 minutes)
  - `15P` (15 minutes)
  - `30P` (30 minutes)
  - `60P` (1 hour)
  - `day` (daily)
  - `week` (weekly)
  - `month` (monthly)
- **Query Parameters**:
  ```
  sdate: DD-MM-YYYY (start date)
  edate: DD-MM-YYYY (end date)
  ```
- **Response Format**:
  ```json
  {
    "data_day": [
      {
        "t": "2024-01-15",
        "o": 85000,
        "h": 86000,
        "l": 84500,
        "c": 85500,
        "v": 1234567,
        "re": 84000,
        "cl": 87000,
        "fl": 81000
      }
    ]
  }
  ```
- **Field Mapping**:
  - `t`: time
  - `o`: open
  - `h`: high
  - `l`: low
  - `c`: close
  - `v`: volume
  - `re`: reference_price
  - `cl`: ceiling_price
  - `fl`: floor_price

#### 2. Intraday Trade History
- **URL**: `GET https://kbbuddywts.kbsec.com.vn/iis-server/investment/trade/history/{symbol}`
- **Query Parameters**:
  ```
  page: 1 (page number)
  limit: 100 (records per page)
  ```
- **Response Format**:
  ```json
  {
    "data": [
      {
        "t": "2026-01-14 14:27:23:15",
        "TD": "14/01/2026",
        "SB": "ACB",
        "FT": "14:27:23",
        "LC": "B",
        "FMP": 25500,
        "FCV": 100,
        "FV": 1000,
        "AVO": 123456,
        "AVA": 3145678900
      }
    ]
  }
  ```
- **Field Mapping**:
  - `t`: timestamp (full)
  - `TD`: trading_date
  - `SB`: symbol
  - `FT`: time
  - `LC`: side (B=buy, S=sell)
  - `FMP`: price
  - `FCV`: price_change
  - `FV`: match_volume
  - `AVO`: accumulated_volume
  - `AVA`: accumulated_value

### Listing Endpoints

#### 3. All Symbols Search
- **URL**: `GET https://kbbuddywts.kbsec.com.vn/iis-server/investment/stock/search/data`
- **Response Format**:
  ```json
  [
    {
      "symbol": "ACB",
      "name": "Ngân hàng TMCP Á Châu",
      "nameEn": "Asia Commercial Bank",
      "exchange": "HOSE",
      "type": "stock",
      "index": 1,
      "re": 25500,
      "ceiling": 26500,
      "floor": 24500
    }
  ]
  ```

#### 4. Symbols by Group/Index
- **URL**: `GET https://kbbuddywts.kbsec.com.vn/iis-server/investment/index/{group_code}/stocks`
- **Group Codes**:
  - `HOSE`: HOSE exchange
  - `HNX`: HNX exchange
  - `UPCOM`: UPCOM exchange
  - `30`: VN30 index
  - `100`: VN100 index
  - `MID`: VNMidCap
  - `SML`: VNSmallCap
  - `SI`: VNSI
  - `X50`: VNX50
  - `XALL`: VNXALL
  - `ALL`: VNALL
  - `HNX30`: HNX30 index
  - `FUND`: ETF/Funds
  - `CW`: Covered Warrants
  - `BOND`: Corporate Bonds
  - `DER`: Derivatives/Futures
- **Response Format**:
  ```json
  {
    "status": 200,
    "data": ["ACB", "VCB", "TCB", "BID"]
  }
  ```

#### 5. Symbols by Industry
- **URL (All Industries)**: `GET https://kbbuddywts.kbsec.com.vn/iis-server/investment/sector/all`
- **URL (Stocks by Industry)**: `GET https://kbbuddywts.kbsec.com.vn/iis-server/investment/sector/stock?code={industry_code}&l=1`
- **Industry Codes**: 1-29 (see const.py for full mapping)
- **Response Format (All Industries)**:
  ```json
  [
    {
      "code": 11,
      "name": "Ngân hàng"
    }
  ]
  ```
- **Response Format (Stocks by Industry)**:
  ```json
  {
    "stocks": [
      {
        "sb": "ACB"
      }
    ]
  }
  ```

### Company Profile Endpoints

#### 6. Company Profile
- **URL**: `GET https://kbbuddywts.kbsec.com.vn/iis-server/investment/stockinfo/profile/{symbol}`
- **Query Parameters**:
  ```
  l: 1 (language: 1=Vietnamese)
  ```
- **Response Format**:
  ```json
  {
    "SM": "Ngân hàng thương mại",
    "SB": "ACB",
    "FD": "1993-05-04",
    "CC": 30000000000000,
    "HM": 15000,
    "LD": "2006-11-08",
    "FV": 10000,
    "EX": "HOSE",
    "LP": 15000,
    "VL": 3000000000,
    "CTP": "Nguyễn Văn A",
    "CTPP": "Chủ tịch HĐQT",
    "ADD": "442 Nguyễn Thị Minh Khai, Q.3, TP.HCM",
    "PHONE": "028-38247247",
    "EMAIL": "info@acb.com.vn",
    "URL": "https://www.acb.com.vn",
    "Leaders": [...],
    "Subsidiaries": [...],
    "Ownership": [...],
    "Shareholders": [...],
    "CharterCapital": [...],
    "LaborStructure": [...]
  }
  ```
- **Field Mapping**: See `_COMPANY_PROFILE_MAP` in const.py

#### 7. Company Events
- **URL**: `GET https://kbbuddywts.kbsec.com.vn/iis-server/investment/stockinfo/event/{symbol}`
- **Query Parameters**:
  ```
  l: 1 (language)
  p: 1 (page)
  s: 10 (page size)
  eID: 1-5 (optional event type filter)
  ```
- **Event Types**:
  - 1: Đại hội cổ đông (Shareholder meeting)
  - 2: Trả cổ tức (Dividend payment)
  - 3: Phát hành (Issuance)
  - 4: Giao dịch cổ đông nội bộ (Insider trading)
  - 5: Sự kiện khác (Other events)

#### 8. Company News
- **URL**: `GET https://kbbuddywts.kbsec.com.vn/iis-server/investment/stockinfo/news/{symbol}`
- **Query Parameters**:
  ```
  l: 1 (language)
  p: 1 (page)
  s: 10 (page size)
  ```

#### 9. Insider Trading
- **URL**: `GET https://kbbuddywts.kbsec.com.vn/iis-server/investment/stockinfo/news/internal-trading/{symbol}`
- **Query Parameters**:
  ```
  l: 1 (language)
  p: 1 (page)
  s: 10 (page size)
  ```

### Financial Statement Endpoints

#### 10. Financial Reports
- **URL**: `GET https://kbbuddywts.kbsec.com.vn/sas/kbsv-stock-data-store/stock/finance-info/{symbol}`
- **Query Parameters**:
  ```
  page: 1
  pageSize: 8
  type: CDKT|KQKD|LCTT|CSTC (report type)
  unit: 1000 (unit in thousands)
  termtype: 1|2 (1=annual, 2=quarterly)
  languageid: 1 (for most reports)
  ```
- **Report Types**:
  - `CDKT`: Balance Sheet (Cân đối kế toán)
  - `KQKD`: Income Statement (Kết quả kinh doanh)
  - `LCTT`: Cash Flow (Lưu chuyển tiền tệ)
  - `CSTC`: Financial Ratios (Chỉ số tài chính)
  - `CTKH`: Planned Indicators
  - `BCTT`: Summary Financial Report
- **Response Format**:
  ```json
  {
    "Audit": [
      {
        "AuditedStatusCode": 1,
        "Description": "Đã kiểm toán"
      }
    ],
    "Unit": ["Triệu đồng"],
    "Head": [
      {
        "YearPeriod": 2024,
        "TermName": "Quý 1",
        "TermNameEN": "Quarter 1",
        "AuditedStatus": 1,
        "ReportDate": "2024-03-31"
      }
    ],
    "Content": {
      "Kết quả kinh doanh": [
        {
          "Name": "Doanh thu",
          "NameEn": "Revenue",
          "Unit": "Triệu đồng",
          "Levels": 1,
          "ID": 1,
          "Value1": 1500000,
          "Value2": 1400000,
          "Value3": 1300000,
          "Value4": 1200000
        }
      ]
    }
  }
  ```

### Index Mapping

**Supported Indices**:
- `VNINDEX`: VN-Index
- `HNXINDEX`: HNX-Index
- `UPCOMINDEX`: UPCOM-Index
- `VN30`: VN30
- `HNX30`: HNX30
- `VN100`: VN100

### Price Scaling

- **OHLC Prices**: Divide by 1000 (API returns in VND * 1000)
- **Volume**: Use as-is (integer)
- **Financial Data**: Based on unit parameter (typically 1000 = thousands VND)

### Rate Limiting

- No explicit rate limits documented
- Recommended: Implement exponential backoff for failed requests
- Use proxy rotation for high-volume requests

### Error Handling

- **Empty Response**: Returns empty array `[]` or empty object `{}`
- **Invalid Symbol**: Returns empty data array
- **Invalid Parameters**: May return HTTP 400 or empty response
- **Server Error**: HTTP 500 with error message

### Implementation Notes

1. **Date Format**: KBS uses DD-MM-YYYY format (different from VCI's YYYY-MM-DD)
2. **Interval Mapping**: Use suffix-based intervals (1P, 5P, day, week, month)
3. **Field Names**: KBS uses short codes (t, o, h, l, c, v) that need mapping
4. **Pagination**: Most endpoints support page/limit or page/pageSize parameters
5. **Language**: Use `l=1` for Vietnamese, language support varies by endpoint
6. **Financial Reports**: Complex nested structure with Audit, Unit, Head, Content sections
7. **Cash Flow**: Has two types (direct/indirect), check both keys in Content
8. **Ratios**: Multiple ratio groups in single response, need to combine them

### Comparison with VCI

| Feature | KBS | VCI |
|---------|-----|-----|
| Date Format | DD-MM-YYYY | YYYY-MM-DD |
| Interval Format | Suffix (1P, day) | Enum (ONE_MINUTE, ONE_DAY) |
| Company Data | 30+ columns | 10 columns |
| Financial Reports | Structured (Audit/Head/Content) | GraphQL |
| Authentication | None | None |
| Stability | High | Medium |
| Data Coverage | Comprehensive | Basic |
