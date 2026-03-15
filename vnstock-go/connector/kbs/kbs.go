// Package kbs implements the Connector interface for the KBS (KB Securities) data source.
package kbs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	vnstock "github.com/dda10/vnstock-go"
)

func init() {
	vnstock.RegisterConnector("KBS", func(client *http.Client, logger *slog.Logger) vnstock.Connector {
		return New(client, logger)
	})
}

const (
	iisBaseURL = "https://kbbuddywts.kbsec.com.vn/iis-server/investment"
	sasBaseURL = "https://kbbuddywts.kbsec.com.vn/sas"
)

// Connector implements the vnstock.Connector interface for KBS data source.
type Connector struct {
	client *http.Client
	logger *slog.Logger
}

// New creates a new KBS connector with the provided HTTP client and logger.
func New(client *http.Client, logger *slog.Logger) *Connector {
	if logger == nil {
		logger = slog.Default()
	}
	return &Connector{
		client: client,
		logger: logger,
	}
}

// logRequest logs an HTTP request at DEBUG level with structured fields.
func (c *Connector) logRequest(method, url string, statusCode int, elapsed time.Duration) {
	c.logger.Debug("KBS request",
		slog.String("method", method),
		slog.String("url", url),
		slog.Int("status", statusCode),
		slog.Duration("elapsed", elapsed),
	)
}

// doRequest performs an HTTP request with optional JSON payload and logs it.
// The url parameter should be the complete URL (not just an endpoint).
func (c *Connector) doRequest(ctx context.Context, method, url string, payload interface{}) (*http.Response, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, &vnstock.Error{
				Code:    vnstock.SerialiseError,
				Message: "failed to marshal request payload",
				Cause:   err,
			}
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.NetworkError,
			Message: "failed to create request",
			Cause:   err,
		}
	}

	// Set standard headers for KBS API
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,vi-VN;q=0.8,vi;q=0.7")
	req.Header.Set("Connection", "keep-alive")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36")

	start := time.Now()
	resp, err := c.client.Do(req)
	elapsed := time.Since(start)

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}
	c.logRequest(method, url, statusCode, elapsed)

	if err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.NetworkError,
			Message: "request failed",
			Cause:   err,
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &vnstock.Error{
			Code:       vnstock.HTTPError,
			Message:    fmt.Sprintf("HTTP error: %s", string(body)),
			StatusCode: resp.StatusCode,
		}
	}

	return resp, nil
}

// validateSymbol checks if a symbol is in valid format (3-10 alphanumeric characters).
func validateSymbol(symbol string) error {
	if symbol == "" {
		return &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: "symbol cannot be empty",
		}
	}

	// Check length
	if len(symbol) < 3 || len(symbol) > 10 {
		return &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: "symbol must be 3-10 characters",
		}
	}

	// Check alphanumeric
	for _, ch := range symbol {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')) {
			return &vnstock.Error{
				Code:    vnstock.InvalidInput,
				Message: "symbol must contain only alphanumeric characters",
			}
		}
	}

	return nil
}

// validateDateRange checks if start date is before end date.
func validateDateRange(start, end time.Time) error {
	if start.After(end) || start.Equal(end) {
		return &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: "start date must be before end date",
		}
	}
	return nil
}

// QuoteHistory retrieves historical OHLCV data for a symbol.
func (c *Connector) QuoteHistory(ctx context.Context, req vnstock.QuoteHistoryRequest) ([]vnstock.Quote, error) {
	// Validate inputs
	if err := validateSymbol(req.Symbol); err != nil {
		return nil, err
	}
	if err := validateDateRange(req.Start, req.End); err != nil {
		return nil, err
	}

	// Map interval to KBS format
	intervalSuffix, err := mapIntervalToKBS(req.Interval)
	if err != nil {
		return nil, err
	}

	// Build URL
	url := fmt.Sprintf("%s/stocks/%s/data_%s?sdate=%s&edate=%s",
		iisBaseURL,
		req.Symbol,
		intervalSuffix,
		req.Start.Format("02-01-2006"),
		req.End.Format("02-01-2006"),
	)

	// Make request
	resp, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	var apiResp kbsQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode response",
			Cause:   err,
		}
	}

	// Check for empty data
	dataKey := "data_" + intervalSuffix
	data, ok := apiResp[dataKey]
	if !ok || len(data) == 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "no data available for the requested date range",
		}
	}

	// Transform to Quote slice
	quotes := make([]vnstock.Quote, 0, len(data))
	for _, item := range data {
		timestamp, err := parseKBSTimestamp(item.Time, intervalSuffix)
		if err != nil {
			c.logger.Warn("failed to parse timestamp", slog.String("time", item.Time), slog.Any("error", err))
			continue
		}

		quotes = append(quotes, vnstock.Quote{
			Symbol:    req.Symbol,
			Timestamp: timestamp,
			Open:      item.Open / 1000.0,
			High:      item.High / 1000.0,
			Low:       item.Low / 1000.0,
			Close:     item.Close / 1000.0,
			Volume:    item.Volume,
			Interval:  req.Interval,
		})
	}

	return quotes, nil
}

// kbsQuoteResponse represents the KBS API response structure.
// The response has a dynamic key like "data_day", "data_1P", etc.
type kbsQuoteResponse map[string][]kbsQuoteItem

// kbsQuoteItem represents a single OHLCV record from KBS API.
type kbsQuoteItem struct {
	Time   string  `json:"t"`
	Open   float64 `json:"o"`
	High   float64 `json:"h"`
	Low    float64 `json:"l"`
	Close  float64 `json:"c"`
	Volume int64   `json:"v"`
}

// mapIntervalToKBS converts vnstock interval format to KBS interval suffix.
func mapIntervalToKBS(interval string) (string, error) {
	mapping := map[string]string{
		"1m":  "1P",
		"5m":  "5P",
		"15m": "15P",
		"30m": "30P",
		"1H":  "60P",
		"1D":  "day",
		"1W":  "week",
		"1M":  "month",
	}

	suffix, ok := mapping[interval]
	if !ok {
		return "", &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: fmt.Sprintf("unsupported interval: %s", interval),
		}
	}
	return suffix, nil
}

// parseKBSTimestamp parses KBS timestamp format based on interval.
// For intraday: "2024-01-15 09:30:00"
// For daily+: "2024-01-15"
func parseKBSTimestamp(timeStr, intervalSuffix string) (time.Time, error) {
	var layout string
	if intervalSuffix == "day" || intervalSuffix == "week" || intervalSuffix == "month" {
		layout = "2006-01-02"
	} else {
		layout = "2006-01-02 15:04:05"
	}

	t, err := time.Parse(layout, timeStr)
	if err != nil {
		return time.Time{}, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse timestamp",
			Cause:   err,
		}
	}
	return t, nil
}

// kbsTradeHistoryResponse represents the KBS intraday trade history API response.
type kbsTradeHistoryResponse struct {
	Data []kbsTradeItem `json:"data"`
}

// kbsTradeItem represents a single trade record from the KBS intraday trade history API.
type kbsTradeItem struct {
	Timestamp        string  `json:"t"`   // full timestamp e.g. "2026-01-14 14:27:23:15"
	TradingDate      string  `json:"TD"`  // e.g. "14/01/2026"
	Symbol           string  `json:"SB"`  // symbol
	Time             string  `json:"FT"`  // e.g. "14:27:23"
	Side             string  `json:"LC"`  // B=buy, S=sell
	Price            float64 `json:"FMP"` // price (divide by 1000)
	PriceChange      float64 `json:"FCV"` // price change
	MatchVolume      int64   `json:"FV"`  // match volume
	AccumulatedVol   int64   `json:"AVO"` // accumulated volume
	AccumulatedValue float64 `json:"AVA"` // accumulated value
}

// RealTimeQuotes retrieves the most recent quote for one or more symbols.
// It fetches the latest trade from the KBS intraday trade history endpoint for each symbol
// and constructs a Quote using the latest trade price as Close and accumulated volume as Volume.
func (c *Connector) RealTimeQuotes(ctx context.Context, symbols []string) ([]vnstock.Quote, error) {
	if len(symbols) == 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: "symbols list cannot be empty",
		}
	}

	// Validate all symbols before making any requests
	for _, symbol := range symbols {
		if err := validateSymbol(symbol); err != nil {
			return nil, err
		}
	}

	quotes := make([]vnstock.Quote, 0, len(symbols))

	for _, symbol := range symbols {
		url := fmt.Sprintf("%s/trade/history/%s?page=1&limit=1", iisBaseURL, symbol)

		resp, err := c.doRequest(ctx, "GET", url, nil)
		if err != nil {
			c.logger.Warn("failed to get real-time quote", slog.String("symbol", symbol), slog.Any("error", err))
			continue
		}

		var tradeResp kbsTradeHistoryResponse
		if err := json.NewDecoder(resp.Body).Decode(&tradeResp); err != nil {
			resp.Body.Close()
			c.logger.Warn("failed to decode trade history response", slog.String("symbol", symbol), slog.Any("error", err))
			continue
		}
		resp.Body.Close()

		if len(tradeResp.Data) == 0 {
			c.logger.Warn("no trade data available", slog.String("symbol", symbol))
			continue
		}

		trade := tradeResp.Data[0]

		// Parse timestamp from TradingDate + Time fields
		timestamp, err := parseTradeTimestamp(trade.TradingDate, trade.Time)
		if err != nil {
			c.logger.Warn("failed to parse trade timestamp", slog.String("symbol", symbol), slog.Any("error", err))
			continue
		}

		price := trade.Price / 1000.0

		quotes = append(quotes, vnstock.Quote{
			Symbol:    symbol,
			Timestamp: timestamp,
			Open:      price,
			High:      price,
			Low:       price,
			Close:     price,
			Volume:    trade.AccumulatedVol,
			Interval:  "realtime",
		})
	}

	if len(quotes) == 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "no real-time data available for any requested symbol",
		}
	}

	return quotes, nil
}

// parseTradeTimestamp parses KBS trade date and time into a time.Time.
// TradingDate format: "DD/MM/YYYY", Time format: "HH:MM:SS"
func parseTradeTimestamp(tradingDate, tradeTime string) (time.Time, error) {
	combined := tradingDate + " " + tradeTime
	t, err := time.Parse("02/01/2006 15:04:05", combined)
	if err != nil {
		return time.Time{}, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to parse trade timestamp",
			Cause:   err,
		}
	}
	return t, nil
}

// kbsListingItem represents a single item from the KBS listing search endpoint.
type kbsListingItem struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	NameEn   string `json:"nameEn"`
	Exchange string `json:"exchange"`
	Type     string `json:"type"`
}

// Listing retrieves the full list of symbols traded on an exchange.
func (c *Connector) Listing(ctx context.Context, exchange string) ([]vnstock.ListingRecord, error) {
	url := iisBaseURL + "/stock/search/data"

	resp, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var items []kbsListingItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode listing response",
			Cause:   err,
		}
	}

	records := make([]vnstock.ListingRecord, 0, len(items))
	for _, item := range items {
		// Filter by exchange if specified
		if exchange != "" && !strings.EqualFold(item.Exchange, exchange) {
			continue
		}

		companyName := item.NameEn
		if companyName == "" {
			companyName = item.Name
		}

		records = append(records, vnstock.ListingRecord{
			Symbol:      item.Symbol,
			Exchange:    item.Exchange,
			CompanyName: companyName,
			Sector:      item.Type,
		})
	}

	if len(records) == 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "no listing data available",
		}
	}

	return records, nil
}

// kbsIndexNameMap maps common index names to KBS API index codes.
var kbsIndexNameMap = map[string]string{
	"VN-Index":    "VNINDEX",
	"HNX-Index":   "HNXINDEX",
	"UPCOM-Index": "UPCOMINDEX",
	"VN30":        "VN30",
	"HNX30":       "HNX30",
	"VN100":       "VN100",
}

// mapIndexName converts a user-facing index name to the KBS API code.
// Returns the mapped name and true if valid, or empty string and false if unrecognized.
func mapIndexName(name string) (string, bool) {
	mapped, ok := kbsIndexNameMap[name]
	return mapped, ok
}

// IndexCurrent retrieves the current value of a named market index.
func (c *Connector) IndexCurrent(ctx context.Context, name string) (vnstock.IndexRecord, error) {
	kbsName, ok := mapIndexName(name)
	if !ok {
		return vnstock.IndexRecord{}, &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: fmt.Sprintf("unrecognized index name: %s", name),
		}
	}

	today := time.Now().Format("02-01-2006")
	url := fmt.Sprintf("%s/index/%s/data_day?sdate=%s&edate=%s", iisBaseURL, kbsName, today, today)

	resp, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return vnstock.IndexRecord{}, err
	}
	defer resp.Body.Close()

	var apiResp kbsQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return vnstock.IndexRecord{}, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode index response",
			Cause:   err,
		}
	}

	data, ok := apiResp["data_day"]
	if !ok || len(data) == 0 {
		return vnstock.IndexRecord{}, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "no index data available for today",
		}
	}

	item := data[len(data)-1] // latest entry
	timestamp, err := parseKBSTimestamp(item.Time, "day")
	if err != nil {
		return vnstock.IndexRecord{}, err
	}

	return vnstock.IndexRecord{
		Name:      name,
		Timestamp: timestamp,
		Value:     item.Close / 1000.0,
		Open:      item.Open / 1000.0,
		High:      item.High / 1000.0,
		Low:       item.Low / 1000.0,
		Close:     item.Close / 1000.0,
		Volume:    item.Volume,
	}, nil
}

// IndexHistory retrieves historical values for a named market index.
func (c *Connector) IndexHistory(ctx context.Context, req vnstock.IndexHistoryRequest) ([]vnstock.IndexRecord, error) {
	kbsName, ok := mapIndexName(req.Name)
	if !ok {
		return nil, &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: fmt.Sprintf("unrecognized index name: %s", req.Name),
		}
	}

	if err := validateDateRange(req.Start, req.End); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/index/%s/data_day?sdate=%s&edate=%s",
		iisBaseURL,
		kbsName,
		req.Start.Format("02-01-2006"),
		req.End.Format("02-01-2006"),
	)

	resp, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp kbsQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode index history response",
			Cause:   err,
		}
	}

	data, ok := apiResp["data_day"]
	if !ok || len(data) == 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "no index history data available for the requested date range",
		}
	}

	records := make([]vnstock.IndexRecord, 0, len(data))
	for _, item := range data {
		timestamp, err := parseKBSTimestamp(item.Time, "day")
		if err != nil {
			c.logger.Warn("failed to parse index timestamp", slog.String("time", item.Time), slog.Any("error", err))
			continue
		}

		records = append(records, vnstock.IndexRecord{
			Name:      req.Name,
			Timestamp: timestamp,
			Value:     item.Close / 1000.0,
			Open:      item.Open / 1000.0,
			High:      item.High / 1000.0,
			Low:       item.Low / 1000.0,
			Close:     item.Close / 1000.0,
			Volume:    item.Volume,
		})
	}

	return records, nil
}

// kbsCompanyProfileResponse represents the KBS company profile API response.
type kbsCompanyProfileResponse struct {
	SB             string                  `json:"SB"`             // symbol
	NM             string                  `json:"NM"`             // company name
	SM             string                  `json:"SM"`             // sector
	IN             string                  `json:"IN"`             // industry
	FD             string                  `json:"FD"`             // founded date
	CC             float64                 `json:"CC"`             // charter capital
	HM             float64                 `json:"HM"`             // market cap
	LD             string                  `json:"LD"`             // listed date
	FV             float64                 `json:"FV"`             // face value
	EX             string                  `json:"EX"`             // exchange
	LP             float64                 `json:"LP"`             // listed price
	VL             int64                   `json:"VL"`             // listed volume
	CTP            string                  `json:"CTP"`            // chairman name
	CTPP           string                  `json:"CTPP"`           // chairman position
	ADD            string                  `json:"ADD"`            // address
	PHONE          string                  `json:"PHONE"`          // phone
	EMAIL          string                  `json:"EMAIL"`          // email
	URL            string                  `json:"URL"`            // website
	DESC           string                  `json:"DESC"`           // description
	Leaders        []kbsLeaderItem         `json:"Leaders"`        // officers list
	Subsidiaries   []kbsSubsidiaryItem     `json:"Subsidiaries"`   // subsidiaries
	Shareholders   []kbsShareholderItem    `json:"Shareholders"`   // major shareholders
	Ownership      []kbsOwnershipItem      `json:"Ownership"`      // ownership structure
	CharterCapital []kbsCharterCapitalItem `json:"CharterCapital"` // charter capital history
	LaborStructure []kbsLaborStructureItem `json:"LaborStructure"` // labor structure
}

// kbsLeaderItem represents a single officer from the KBS company profile response.
type kbsLeaderItem struct {
	Name  string `json:"Name"`
	Title string `json:"Title"`
}

// kbsSubsidiaryItem represents a subsidiary from the KBS company profile response.
type kbsSubsidiaryItem struct {
	Name      string  `json:"Name"`
	Ownership float64 `json:"Ownership"`
	Capital   float64 `json:"Capital"`
	Note      string  `json:"Note"`
}

// kbsShareholderItem represents a shareholder from the KBS company profile response.
type kbsShareholderItem struct {
	Name       string  `json:"Name"`
	Shares     float64 `json:"Shares"`
	Percentage float64 `json:"Percentage"`
	Note       string  `json:"Note"`
}

// kbsOwnershipItem represents an ownership entry from the KBS company profile response.
type kbsOwnershipItem struct {
	Name       string  `json:"Name"`
	Shares     float64 `json:"Shares"`
	Percentage float64 `json:"Percentage"`
}

// kbsCharterCapitalItem represents a charter capital record from the KBS company profile response.
type kbsCharterCapitalItem struct {
	Date   string  `json:"Date"`
	Amount float64 `json:"Amount"`
	Note   string  `json:"Note"`
}

// kbsLaborStructureItem represents a labor structure entry from the KBS company profile response.
type kbsLaborStructureItem struct {
	Year      int   `json:"Year"`
	Employees int64 `json:"Employees"`
}

// CompanyProfile retrieves descriptive information about a listed company.
func (c *Connector) CompanyProfile(ctx context.Context, symbol string) (vnstock.CompanyProfile, error) {
	if err := validateSymbol(symbol); err != nil {
		return vnstock.CompanyProfile{}, err
	}

	url := fmt.Sprintf("%s/stockinfo/profile/%s?l=1", iisBaseURL, symbol)

	resp, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return vnstock.CompanyProfile{}, err
	}
	defer resp.Body.Close()

	var profile kbsCompanyProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return vnstock.CompanyProfile{}, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode company profile response",
			Cause:   err,
		}
	}

	// Use NM field for name, fall back to symbol
	companyName := profile.NM
	if companyName == "" {
		companyName = profile.SB
	}

	// Map leaders
	leaders := make([]vnstock.Officer, 0, len(profile.Leaders))
	for _, l := range profile.Leaders {
		leaders = append(leaders, vnstock.Officer{
			Name:  l.Name,
			Title: l.Title,
		})
	}

	// Map subsidiaries
	subsidiaries := make([]vnstock.Subsidiary, 0, len(profile.Subsidiaries))
	for _, s := range profile.Subsidiaries {
		subsidiaries = append(subsidiaries, vnstock.Subsidiary{
			Name:      s.Name,
			Ownership: s.Ownership,
			Capital:   s.Capital,
			Note:      s.Note,
		})
	}

	// Map shareholders
	shareholders := make([]vnstock.Shareholder, 0, len(profile.Shareholders))
	for _, s := range profile.Shareholders {
		shareholders = append(shareholders, vnstock.Shareholder{
			Name:       s.Name,
			Shares:     s.Shares,
			Percentage: s.Percentage,
			Note:       s.Note,
		})
	}

	// Map ownership
	ownership := make([]vnstock.OwnershipEntry, 0, len(profile.Ownership))
	for _, o := range profile.Ownership {
		ownership = append(ownership, vnstock.OwnershipEntry{
			Name:       o.Name,
			Shares:     o.Shares,
			Percentage: o.Percentage,
		})
	}

	// Map charter capital history
	charterHistory := make([]vnstock.CharterCapitalRec, 0, len(profile.CharterCapital))
	for _, ch := range profile.CharterCapital {
		charterHistory = append(charterHistory, vnstock.CharterCapitalRec{
			Date:   ch.Date,
			Amount: ch.Amount,
			Note:   ch.Note,
		})
	}

	// Map labor structure
	laborStructure := make([]vnstock.LaborEntry, 0, len(profile.LaborStructure))
	for _, lb := range profile.LaborStructure {
		laborStructure = append(laborStructure, vnstock.LaborEntry{
			Year:      lb.Year,
			Employees: lb.Employees,
		})
	}

	return vnstock.CompanyProfile{
		Symbol:           profile.SB,
		Name:             companyName,
		Exchange:         profile.EX,
		Sector:           profile.SM,
		Industry:         profile.IN,
		Founded:          profile.FD,
		Website:          profile.URL,
		Description:      profile.DESC,
		Address:          profile.ADD,
		Phone:            profile.PHONE,
		Email:            profile.EMAIL,
		CharterCapital:   profile.CC,
		ListedDate:       profile.LD,
		FaceValue:        profile.FV,
		ListedPrice:      profile.LP,
		ListedVolume:     profile.VL,
		MarketCap:        profile.HM,
		ChairmanName:     profile.CTP,
		ChairmanPosition: profile.CTPP,
		Leaders:          leaders,
		Subsidiaries:     subsidiaries,
		Shareholders:     shareholders,
		Ownership:        ownership,
		CharterHistory:   charterHistory,
		LaborStructure:   laborStructure,
	}, nil
}

// Officers retrieves the list of officers and executives for a company.
func (c *Connector) Officers(ctx context.Context, symbol string) ([]vnstock.Officer, error) {
	if err := validateSymbol(symbol); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/stockinfo/profile/%s?l=1", iisBaseURL, symbol)

	resp, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var profile kbsCompanyProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode officers response",
			Cause:   err,
		}
	}

	if len(profile.Leaders) == 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "no officers data available for symbol",
		}
	}

	officers := make([]vnstock.Officer, 0, len(profile.Leaders))
	for _, leader := range profile.Leaders {
		officers = append(officers, vnstock.Officer{
			Name:  leader.Name,
			Title: leader.Title,
		})
	}

	return officers, nil
}

// kbsStatementTypeMap maps user-facing statement types to KBS API type codes.
var kbsStatementTypeMap = map[string]string{
	"income":   "KQKD",
	"balance":  "CDKT",
	"cashflow": "LCTT",
}

// kbsFinancialResponse represents the KBS financial statement API response.
type kbsFinancialResponse struct {
	Head    []kbsFinancialHead            `json:"Head"`
	Content map[string][]kbsFinancialItem `json:"Content"`
}

// kbsFinancialHead represents a period header in the financial response.
type kbsFinancialHead struct {
	YearPeriod int    `json:"YearPeriod"`
	TermName   string `json:"TermName"`
	ReportDate string `json:"ReportDate"`
}

// kbsFinancialItem represents a single financial line item.
type kbsFinancialItem struct {
	Name   string   `json:"Name"`
	NameEn string   `json:"NameEn"`
	Value1 *float64 `json:"Value1"`
	Value2 *float64 `json:"Value2"`
	Value3 *float64 `json:"Value3"`
	Value4 *float64 `json:"Value4"`
	Value5 *float64 `json:"Value5"`
	Value6 *float64 `json:"Value6"`
	Value7 *float64 `json:"Value7"`
	Value8 *float64 `json:"Value8"`
}

// getValues returns the Value1..Value8 fields as a slice of *float64.
func (item *kbsFinancialItem) getValues() []*float64 {
	return []*float64{item.Value1, item.Value2, item.Value3, item.Value4, item.Value5, item.Value6, item.Value7, item.Value8}
}

// parseQuarterFromTermName extracts the quarter number from a KBS TermName string.
// e.g. "Quý 1" -> 1, "Quarter 2" -> 2, "Năm" -> 0 (annual)
func parseQuarterFromTermName(termName string) int {
	termName = strings.TrimSpace(termName)
	// Try to find a digit in the term name
	for _, ch := range termName {
		if ch >= '1' && ch <= '4' {
			return int(ch - '0')
		}
	}
	return 0
}

// FinancialStatement retrieves financial statement data for a company.
func (c *Connector) FinancialStatement(ctx context.Context, req vnstock.FinancialRequest) ([]vnstock.FinancialPeriod, error) {
	if err := validateSymbol(req.Symbol); err != nil {
		return nil, err
	}

	kbsType, ok := kbsStatementTypeMap[req.Type]
	if !ok {
		return nil, &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: fmt.Sprintf("invalid statement type: %s (must be income, balance, or cashflow)", req.Type),
		}
	}

	termType := "1" // annual
	if req.Period == "quarterly" {
		termType = "2"
	} else if req.Period != "" && req.Period != "annual" {
		return nil, &vnstock.Error{
			Code:    vnstock.InvalidInput,
			Message: fmt.Sprintf("invalid period: %s (must be annual or quarterly)", req.Period),
		}
	}

	url := fmt.Sprintf("%s/kbsv-stock-data-store/stock/finance-info/%s?page=1&pageSize=8&type=%s&unit=1000&termtype=%s&languageid=1",
		sasBaseURL,
		req.Symbol,
		kbsType,
		termType,
	)

	resp, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var finResp kbsFinancialResponse
	if err := json.NewDecoder(resp.Body).Decode(&finResp); err != nil {
		return nil, &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "failed to decode financial statement response",
			Cause:   err,
		}
	}

	if len(finResp.Head) == 0 {
		return nil, &vnstock.Error{
			Code:    vnstock.NoData,
			Message: "no financial data available",
		}
	}

	// Build a FinancialPeriod for each Head entry
	numPeriods := len(finResp.Head)
	periods := make([]vnstock.FinancialPeriod, numPeriods)
	for i, head := range finResp.Head {
		quarter := parseQuarterFromTermName(head.TermName)
		period := req.Period
		if period == "" {
			period = "annual"
		}

		periods[i] = vnstock.FinancialPeriod{
			Symbol:  req.Symbol,
			Period:  period,
			Year:    head.YearPeriod,
			Quarter: quarter,
			Fields:  make(map[string]float64),
		}
	}

	// Populate Fields from Content sections
	for _, items := range finResp.Content {
		for _, item := range items {
			fieldName := item.NameEn
			if fieldName == "" {
				fieldName = item.Name
			}
			if fieldName == "" {
				continue
			}

			values := item.getValues()
			for i := 0; i < numPeriods && i < len(values); i++ {
				if values[i] != nil {
					periods[i].Fields[fieldName] = *values[i]
				}
			}
		}
	}

	// Sort by (Year, Quarter) descending
	sort.Slice(periods, func(i, j int) bool {
		if periods[i].Year != periods[j].Year {
			return periods[i].Year > periods[j].Year
		}
		return periods[i].Quarter > periods[j].Quarter
	})

	return periods, nil
}
