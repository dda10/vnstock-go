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
