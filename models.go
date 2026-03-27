package vnstock

import "time"

// Quote represents a price record for a symbol at a point in time.
type Quote struct {
	Symbol    string    `json:"symbol"`
	Timestamp time.Time `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    int64     `json:"volume"`
	Interval  string    `json:"interval"`
}

// ListingRecord represents a listed security on an exchange.
type ListingRecord struct {
	Symbol      string `json:"symbol"`
	Exchange    string `json:"exchange"`
	CompanyName string `json:"company_name"`
	Sector      string `json:"sector"`
}

// IndexRecord represents a market index value at a point in time.
type IndexRecord struct {
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Change    float64   `json:"change"`
	ChangePct float64   `json:"change_pct"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    int64     `json:"volume"`
}

// CompanyProfile contains descriptive information about a listed company.
type CompanyProfile struct {
	Symbol           string              `json:"symbol"`
	Name             string              `json:"name"`
	Exchange         string              `json:"exchange"`
	Sector           string              `json:"sector"`
	Industry         string              `json:"industry"`
	Founded          string              `json:"founded"`
	Website          string              `json:"website"`
	Description      string              `json:"description"`
	Address          string              `json:"address,omitempty"`
	Phone            string              `json:"phone,omitempty"`
	Email            string              `json:"email,omitempty"`
	CharterCapital   float64             `json:"charter_capital,omitempty"`
	ListedDate       string              `json:"listed_date,omitempty"`
	FaceValue        float64             `json:"face_value,omitempty"`
	ListedPrice      float64             `json:"listed_price,omitempty"`
	ListedVolume     int64               `json:"listed_volume,omitempty"`
	MarketCap        float64             `json:"market_cap,omitempty"`
	ChairmanName     string              `json:"chairman_name,omitempty"`
	ChairmanPosition string              `json:"chairman_position,omitempty"`
	Leaders          []Officer           `json:"leaders,omitempty"`
	Subsidiaries     []Subsidiary        `json:"subsidiaries,omitempty"`
	Shareholders     []Shareholder       `json:"shareholders,omitempty"`
	Ownership        []OwnershipEntry    `json:"ownership,omitempty"`
	CharterHistory   []CharterCapitalRec `json:"charter_history,omitempty"`
	LaborStructure   []LaborEntry        `json:"labor_structure,omitempty"`
}

// Subsidiary represents a subsidiary or affiliated company.
type Subsidiary struct {
	Name      string  `json:"name"`
	Ownership float64 `json:"ownership"`
	Capital   float64 `json:"capital,omitempty"`
	Note      string  `json:"note,omitempty"`
}

// Shareholder represents a major shareholder of a company.
type Shareholder struct {
	Name       string  `json:"name"`
	Shares     float64 `json:"shares"`
	Percentage float64 `json:"percentage"`
	Note       string  `json:"note,omitempty"`
}

// OwnershipEntry represents an ownership structure entry.
type OwnershipEntry struct {
	Name       string  `json:"name"`
	Shares     float64 `json:"shares"`
	Percentage float64 `json:"percentage"`
}

// CharterCapitalRec represents a charter capital change record.
type CharterCapitalRec struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
	Note   string  `json:"note,omitempty"`
}

// LaborEntry represents a labor structure entry.
type LaborEntry struct {
	Year      int   `json:"year"`
	Employees int64 `json:"employees"`
}

// Officer represents a company officer or executive.
type Officer struct {
	Name            string `json:"name"`
	Title           string `json:"title"`
	AppointmentDate string `json:"appointment_date"`
}

// FinancialPeriod represents financial data for a specific period.
type FinancialPeriod struct {
	Symbol  string             `json:"symbol"`
	Period  string             `json:"period"` // "annual" or "quarterly"
	Year    int                `json:"year"`
	Quarter int                `json:"quarter"` // 0 for annual
	Fields  map[string]float64 `json:"fields"`
}

// QuoteHistoryRequest specifies parameters for historical quote retrieval.
type QuoteHistoryRequest struct {
	Symbol   string    `json:"symbol"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Interval string    `json:"interval"`
}

// IndexHistoryRequest specifies parameters for historical index retrieval.
type IndexHistoryRequest struct {
	Name     string    `json:"name"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Interval string    `json:"interval"`
}

// FinancialRequest specifies parameters for financial statement retrieval.
type FinancialRequest struct {
	Symbol string `json:"symbol"`
	Type   string `json:"type"`   // "income", "balance", "cashflow"
	Period string `json:"period"` // "annual", "quarterly"
}

// GoldPrice represents gold price data from various sources.
type GoldPrice struct {
	TypeName  string    `json:"type_name"`  // Gold type (e.g., "SJC 1L", "SJC 5C", "24K")
	Branch    string    `json:"branch"`     // Branch name (for SJC)
	BuyPrice  float64   `json:"buy_price"`  // Buy price in thousands VND per tael
	SellPrice float64   `json:"sell_price"` // Sell price in thousands VND per tael
	Date      time.Time `json:"date"`       // Price date
	Source    string    `json:"source"`     // Data source ("SJC", "BTMC")
}

// GoldPriceRequest specifies parameters for gold price retrieval.
type GoldPriceRequest struct {
	Date   time.Time `json:"date"`   // Date for historical prices (optional, defaults to today)
	Source string    `json:"source"` // "SJC" or "BTMC" (optional, defaults to all sources)
}

// CompanyEvent represents a corporate event (dividends, AGM, etc.).
type CompanyEvent struct {
	Symbol      string    `json:"symbol"`
	EventType   string    `json:"event_type"`   // "dividend", "agm", "rights", etc.
	Title       string    `json:"title"`        // Event title/description
	ExDate      time.Time `json:"ex_date"`      // Ex-date for the event
	RecordDate  time.Time `json:"record_date"`  // Record date
	PaymentDate time.Time `json:"payment_date"` // Payment/execution date
	Content     string    `json:"content"`      // Additional details
	Value       float64   `json:"value"`        // Dividend amount or ratio
}

// CompanyNews represents a news article about a company.
type CompanyNews struct {
	Symbol      string    `json:"symbol"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"published_at"`
	URL         string    `json:"url"`
}

// InsiderTrade represents an insider trading transaction.
type InsiderTrade struct {
	Symbol          string    `json:"symbol"`
	InsiderName     string    `json:"insider_name"`
	Position        string    `json:"position"`         // Position/title of insider
	TransactionType string    `json:"transaction_type"` // "buy" or "sell"
	Shares          int64     `json:"shares"`           // Number of shares
	Price           float64   `json:"price"`            // Transaction price
	Value           float64   `json:"value"`            // Total transaction value
	SharesBefore    int64     `json:"shares_before"`    // Shares held before
	SharesAfter     int64     `json:"shares_after"`     // Shares held after
	TransactionDate time.Time `json:"transaction_date"`
	ReportDate      time.Time `json:"report_date"`
}

// SymbolGroup represents a group of symbols (index constituents, industry, etc.).
type SymbolGroup struct {
	GroupCode   string   `json:"group_code"`
	GroupName   string   `json:"group_name"`
	Description string   `json:"description"`
	Symbols     []string `json:"symbols"`
}

// IndustryInfo represents an industry classification with its symbols.
type IndustryInfo struct {
	IndustryCode string   `json:"industry_code"`
	IndustryName string   `json:"industry_name"`
	Symbols      []string `json:"symbols"`
}

// FinancialRatio represents key financial ratios for a company.
// Ported from vnquant's basic index functionality.
type FinancialRatio struct {
	Symbol            string    `json:"symbol"`
	ReportDate        time.Time `json:"report_date"`
	ROA               float64   `json:"roa"`                  // Return on Assets (last 4 quarters)
	ROE               float64   `json:"roe"`                  // Return on Equity (last 4 quarters)
	NetProfitMargin   float64   `json:"net_profit_margin"`    // Net Profit Margin (yearly)
	RevenueGrowth     float64   `json:"revenue_growth"`       // Net Revenue Growth YoY
	ProfitGrowth      float64   `json:"profit_growth"`        // Profit After Tax Growth YoY
	EPS               float64   `json:"eps"`                  // Earnings Per Share
	PE                float64   `json:"pe"`                   // Price to Earnings ratio
	PB                float64   `json:"pb"`                   // Price to Book ratio
	CurrentRatio      float64   `json:"current_ratio"`        // Current Assets / Current Liabilities
	DebtToEquity      float64   `json:"debt_to_equity"`       // Total Debt / Equity
	DividendYield     float64   `json:"dividend_yield"`       // Dividend Yield %
	BookValuePerShare float64   `json:"book_value_per_share"` // Book Value Per Share
}

// FinancialRatioRequest specifies parameters for financial ratio retrieval.
type FinancialRatioRequest struct {
	Symbol     string    `json:"symbol"`
	ReportDate time.Time `json:"report_date"` // Optional, defaults to latest
}
