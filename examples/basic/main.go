package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dda10/vnstock-go"
	_ "github.com/dda10/vnstock-go/all" // Register all connectors
)

func main() {
	// Create a client with VCI connector
	cfg := vnstock.Config{
		Connector: "VCI",
		Timeout:   30 * time.Second,
	}

	client, err := vnstock.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("✓ Client created successfully with VCI connector")

	// Example 1: Get historical quotes
	fmt.Println("\n--- Example 1: Historical Quotes ---")
	req := vnstock.QuoteHistoryRequest{
		Symbol:   "VNM",
		Start:    time.Now().AddDate(0, -1, 0), // 1 month ago
		End:      time.Now(),
		Interval: "1d",
	}

	quotes, err := client.QuoteHistory(context.Background(), req)
	if err != nil {
		fmt.Printf("Error getting quotes: %v\n", err)
	} else {
		fmt.Printf("Retrieved %d quotes for VNM\n", len(quotes))
		if len(quotes) > 0 {
			fmt.Printf("Latest quote: %+v\n", quotes[len(quotes)-1])
		}
	}

	// Example 2: Get real-time quotes
	fmt.Println("\n--- Example 2: Real-Time Quotes ---")
	symbols := []string{"VNM", "HPG", "VIC"}
	rtQuotes, err := client.RealTimeQuotes(context.Background(), symbols)
	if err != nil {
		fmt.Printf("Error getting real-time quotes: %v\n", err)
	} else {
		fmt.Printf("Retrieved %d real-time quotes\n", len(rtQuotes))
		for _, q := range rtQuotes {
			fmt.Printf("  %s: Close=%.2f, Volume=%d\n", q.Symbol, q.Close, q.Volume)
		}
	}

	// Example 3: Get market listing
	fmt.Println("\n--- Example 3: Market Listing ---")
	listing, err := client.Listing(context.Background(), "HOSE")
	if err != nil {
		fmt.Printf("Error getting listing: %v\n", err)
	} else {
		fmt.Printf("Retrieved %d listings from HOSE\n", len(listing))
		if len(listing) > 0 {
			fmt.Printf("First listing: %+v\n", listing[0])
		}
	}

	// Example 4: Get company profile
	fmt.Println("\n--- Example 4: Company Profile ---")
	profile, err := client.CompanyProfile(context.Background(), "VNM")
	if err != nil {
		fmt.Printf("Error getting company profile: %v\n", err)
	} else {
		fmt.Printf("Company: %s\n", profile.Name)
		fmt.Printf("Sector: %s\n", profile.Sector)
		fmt.Printf("Exchange: %s\n", profile.Exchange)
	}

	fmt.Println("\n✓ All examples completed!")
}
