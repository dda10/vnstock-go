# vnstock-go Feature Coverage

This document maps the features from the Python vnstock library to the Go implementation.

## Legend
- ✅ Implemented
- ⚠️ Partially Implemented
- ❌ Not Implemented
- 🔄 In Progress

## 1. Quote Data (Historical & Real-time)

| Feature | Python | Go Status | Notes |
|---------|--------|-----------|-------|
| Historical OHLC | ✅ | ✅ | VCI connector working with real API |
| Real-time quotes | ✅ | ✅ | VCI connector working with real API |
| Intraday data | ✅ | ❌ | Not implemented |
| Multiple intervals | ✅ | ⚠️ | Basic support, needs testing |

## 2. Company Information

| Feature | Python (KBS) | Python (VCI) | Go Status | Notes |
|---------|--------------|--------------|-----------|-------|
| Company overview | ✅ (30 cols) | ✅ (10 cols) | ❌ | Not implemented |
| Shareholders | ✅ (4 cols) | ✅ (5 cols) | ❌ | Not implemented |
| Officers | ✅ (5 cols) | ✅ (7 cols) | ❌ | Interface exists, not implemented |
| Subsidiaries | ✅ (6 cols) | ❌ | ❌ | Not implemented |
| Affiliates | ✅ (6 cols) | ✅ (4 cols) | ❌ | Not implemented |
| News | ✅ (5 cols) | ✅ (18 cols) | ❌ | Not implemented |
| Events | ❌ (empty) | ✅ (13 cols) | ❌ | Not implemented |
| Ownership structure | ✅ (4 cols) | ❌ | ❌ | Not implemented |
| Capital history | ✅ (3 cols) | ❌ | ❌ | Not implemented |
| Insider trading | ✅ | ❌ | ❌ | Not implemented |
| Reports | ❌ | ✅ | ❌ | Not implemented |
| Trading stats | ❌ | ✅ (24 cols) | ❌ | Not implemented |
| Ratio summary | ❌ | ✅ (46 cols) | ❌ | Not implemented |


## 3. Listing Information

| Feature | Python | Go Status | Notes |
|---------|--------|-----------|-------|
| Stock listing | ✅ | ⚠️ | Interface exists, marked NOT_SUPPORTED |
| Exchange filtering | ✅ | ❌ | Not implemented |
| Industry classification | ✅ | ❌ | Not implemented |

## 4. Financial Statements

| Feature | Python | Go Status | Notes |
|---------|--------|-----------|-------|
| Balance sheet | ✅ | ⚠️ | Interface exists, marked NOT_SUPPORTED |
| Income statement | ✅ | ⚠️ | Interface exists, marked NOT_SUPPORTED |
| Cash flow | ✅ | ⚠️ | Interface exists, marked NOT_SUPPORTED |
| Financial ratios | ✅ | ❌ | Not implemented |

## 5. Market Indices

| Feature | Python | Go Status | Notes |
|---------|--------|-----------|-------|
| Index current value | ✅ | ⚠️ | Interface exists, marked NOT_SUPPORTED |
| Index history | ✅ | ⚠️ | Interface exists, marked NOT_SUPPORTED |
| Index components | ✅ | ❌ | Not implemented |

## 6. Trading Data

| Feature | Python | Go Status | Notes |
|---------|--------|-----------|-------|
| Price board | ✅ | ❌ | Not implemented |
| Order book | ✅ | ❌ | Not implemented |
| Trading statistics | ✅ | ❌ | Not implemented |
| Foreign trading | ✅ | ❌ | Not implemented |

## 7. Stock Screener

| Feature | Python | Go Status | Notes |
|---------|--------|-----------|-------|
| Technical filters | ✅ | ❌ | Not implemented |
| Fundamental filters | ✅ | ❌ | Not implemented |
| Custom criteria | ✅ | ❌ | Not implemented |


## 8. Mutual Funds

| Feature | Python | Go Status | Notes |
|---------|--------|-----------|-------|
| Fund listing | ✅ | ❌ | Not implemented |
| Fund NAV history | ✅ | ❌ | Not implemented |
| Fund portfolio | ✅ | ❌ | Not implemented |

## 9. Commodities & Forex

| Feature | Python | Go Status | Notes |
|---------|--------|-----------|-------|
| Gold prices | ✅ | ❌ | Not implemented |
| Oil prices | ✅ | ❌ | Not implemented |
| Forex rates | ✅ | ❌ | Not implemented |
| Cryptocurrency | ✅ | ❌ | Binance connector stub exists |

## 10. Data Visualization

| Feature | Python | Go Status | Notes |
|---------|--------|-----------|-------|
| Chart plotting | ✅ | ❌ | Not applicable (Go library) |
| Technical indicators | ✅ | ❌ | Not implemented |

## 11. Export Functionality

| Feature | Python | Go Status | Notes |
|---------|--------|-----------|-------|
| CSV export | ✅ | ✅ | Implemented |
| JSON export | ✅ | ✅ | Implemented |
| Excel export | ✅ | ✅ | Implemented with excelize |

## 12. Messaging & Notifications

| Feature | Python | Go Status | Notes |
|---------|--------|-----------|-------|
| Telegram bot | ✅ | ❌ | Not implemented |
| Slack integration | ✅ | ❌ | Not implemented |
| Lark integration | ✅ | ❌ | Not implemented |

## 13. Trading API

| Feature | Python | Go Status | Notes |
|---------|--------|-----------|-------|
| DNSE trading API | ✅ | ❌ | Not implemented |
| Order placement | ✅ | ❌ | Not implemented |
| Portfolio management | ✅ | ❌ | Not implemented |

