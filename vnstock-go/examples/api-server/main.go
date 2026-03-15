package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dda10/vnstock-go"
	_ "github.com/dda10/vnstock-go/all"
)

var client *vnstock.Client

func main() {
	// Initialize vnstock client
	cfg := vnstock.Config{
		Connector: "VCI",
		Timeout:   30 * time.Second,
	}

	var err error
	client, err = vnstock.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Register routes
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/quote/history", quoteHistoryHandler)
	http.HandleFunc("/api/quote/realtime", realTimeQuotesHandler)
	http.HandleFunc("/api/listing", listingHandler)
	http.HandleFunc("/api/company/profile", companyProfileHandler)
	http.HandleFunc("/api/index/current", indexCurrentHandler)
	http.HandleFunc("/api/index/history", indexHistoryHandler)

	// Start server
	port := ":8080"
	fmt.Printf("🚀 API Server running on http://localhost%s\n", port)
	fmt.Println("📝 Using VCI connector")
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  GET  /health")
	fmt.Println("  GET  /api/quote/history?symbol=VNM&days=30")
	fmt.Println("  GET  /api/quote/realtime?symbols=VNM,HPG,VIC")
	fmt.Println("  GET  /api/listing?exchange=HOSE")
	fmt.Println("  GET  /api/company/profile?symbol=VNM")
	fmt.Println("  GET  /api/index/current?name=VNINDEX")
	fmt.Println("  GET  /api/index/history?name=VNINDEX&days=30")
	fmt.Println()

	log.Fatal(http.ListenAndServe(port, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "ok",
		"service":   "vnstock-go API",
		"connector": "VCI",
	})
}

func quoteHistoryHandler(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "symbol parameter is required", http.StatusBadRequest)
		return
	}

	days := 30
	fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)

	req := vnstock.QuoteHistoryRequest{
		Symbol:   symbol,
		Start:    time.Now().AddDate(0, 0, -days),
		End:      time.Now(),
		Interval: "1d",
	}

	quotes, err := client.QuoteHistory(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quotes)
}

func realTimeQuotesHandler(w http.ResponseWriter, r *http.Request) {
	symbolsParam := r.URL.Query().Get("symbols")
	if symbolsParam == "" {
		http.Error(w, "symbols parameter is required (comma-separated)", http.StatusBadRequest)
		return
	}

	symbols := parseSymbols(symbolsParam)

	quotes, err := client.RealTimeQuotes(context.Background(), symbols)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quotes)
}

func listingHandler(w http.ResponseWriter, r *http.Request) {
	exchange := r.URL.Query().Get("exchange")

	listing, err := client.Listing(context.Background(), exchange)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listing)
}

func companyProfileHandler(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "symbol parameter is required", http.StatusBadRequest)
		return
	}

	profile, err := client.CompanyProfile(context.Background(), symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func indexCurrentHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name parameter is required", http.StatusBadRequest)
		return
	}

	record, err := client.IndexCurrent(context.Background(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

func indexHistoryHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name parameter is required", http.StatusBadRequest)
		return
	}

	days := 30
	fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days)

	req := vnstock.IndexHistoryRequest{
		Name:  name,
		Start: time.Now().AddDate(0, 0, -days),
		End:   time.Now(),
	}

	records, err := client.IndexHistory(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

func parseSymbols(s string) []string {
	var symbols []string
	current := ""
	for _, c := range s {
		if c == ',' {
			if current != "" {
				symbols = append(symbols, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		symbols = append(symbols, current)
	}
	return symbols
}
