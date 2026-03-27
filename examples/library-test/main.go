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
	fmt.Println("=== Vnstock-Go Library Structure Test ===")

	// Test 1: Client Creation
	fmt.Println("Test 1: Client Creation")
	testClientCreation()

	// Test 2: Config Validation
	fmt.Println("\nTest 2: Config Validation")
	testConfigValidation()

	// Test 3: Connector Registry
	fmt.Println("\nTest 3: Connector Registry")
	testConnectorRegistry()

	// Test 4: Error Handling
	fmt.Println("\nTest 4: Error Handling")
	testErrorHandling()

	// Test 5: Date Validation
	fmt.Println("\nTest 5: Date Validation")
	testDateValidation()

	fmt.Println("\n=== All Library Structure Tests Passed! ===")
	fmt.Println("\nNote: These tests validate the library structure works correctly.")
	fmt.Println("To test with real data, you need to:")
	fmt.Println("1. Update connector implementations with real API endpoints")
	fmt.Println("2. Set useMockData=false in the API server")
	fmt.Println("3. Test with Postman or curl")
}

func testClientCreation() {
	cfg := vnstock.Config{
		Connector: "VCI",
		Timeout:   30 * time.Second,
	}

	client, err := vnstock.New(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	if client == nil {
		log.Fatal("❌ Client is nil")
	}

	fmt.Println("✅ Client created successfully")
}

func testConfigValidation() {
	// Test 1: Empty connector
	cfg := vnstock.Config{
		Timeout: 30 * time.Second,
	}
	_, err := vnstock.New(cfg)
	if err == nil {
		log.Fatal("❌ Should fail with empty connector")
	}
	fmt.Println("✅ Empty connector validation works")

	// Test 2: Invalid connector
	cfg = vnstock.Config{
		Connector: "INVALID",
		Timeout:   30 * time.Second,
	}
	_, err = vnstock.New(cfg)
	if err == nil {
		log.Fatal("❌ Should fail with invalid connector")
	}
	fmt.Println("✅ Invalid connector validation works")

	// Test 3: Negative timeout
	cfg = vnstock.Config{
		Connector: "VCI",
		Timeout:   -1 * time.Second,
	}
	_, err = vnstock.New(cfg)
	if err == nil {
		log.Fatal("❌ Should fail with negative timeout")
	}
	fmt.Println("✅ Negative timeout validation works")
}

func testConnectorRegistry() {
	connectors := []string{"VCI", "DNSE", "FMP", "Binance"}

	for _, name := range connectors {
		cfg := vnstock.Config{
			Connector: name,
			Timeout:   30 * time.Second,
		}

		client, err := vnstock.New(cfg)
		if err != nil {
			log.Fatalf("❌ Failed to create %s connector: %v", name, err)
		}

		if client == nil {
			log.Fatalf("❌ %s connector returned nil client", name)
		}

		fmt.Printf("✅ %s connector registered and working\n", name)
	}
}

func testErrorHandling() {
	cfg := vnstock.Config{
		Connector: "VCI",
		Timeout:   30 * time.Second,
	}

	client, err := vnstock.New(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	// Test invalid date range
	req := vnstock.QuoteHistoryRequest{
		Symbol:   "VNM",
		Start:    time.Now(),
		End:      time.Now().AddDate(0, 0, -30), // End before Start
		Interval: "1d",
	}

	_, err = client.QuoteHistory(context.Background(), req)
	if err == nil {
		log.Fatal("❌ Should fail with invalid date range")
	}

	// Check if it's the right error type
	if vnstockErr, ok := err.(*vnstock.Error); ok {
		if vnstockErr.Code != vnstock.InvalidInput {
			log.Fatalf("❌ Wrong error code: got %s, want %s", vnstockErr.Code, vnstock.InvalidInput)
		}
		fmt.Println("✅ Date range validation works")
	} else {
		log.Fatal("❌ Error is not *vnstock.Error type")
	}
}

func testDateValidation() {
	cfg := vnstock.Config{
		Connector: "VCI",
		Timeout:   30 * time.Second,
	}

	client, err := vnstock.New(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	// Test QuoteHistory date validation
	req := vnstock.QuoteHistoryRequest{
		Symbol:   "VNM",
		Start:    time.Now().AddDate(0, 0, -30),
		End:      time.Now(),
		Interval: "1d",
	}

	// This will fail at the connector level (no real API), but should pass client-level validation
	_, err = client.QuoteHistory(context.Background(), req)

	// We expect a network error (because API doesn't exist), not an InvalidInput error
	if err != nil {
		if vnstockErr, ok := err.(*vnstock.Error); ok {
			if vnstockErr.Code == vnstock.InvalidInput {
				log.Fatal("❌ Date validation failed - should accept valid date range")
			}
		}
	}

	fmt.Println("✅ Valid date range accepted by client")

	// Test IndexHistory date validation
	indexReq := vnstock.IndexHistoryRequest{
		Name:  "VNINDEX",
		Start: time.Now().AddDate(0, 0, -30),
		End:   time.Now(),
	}

	_, err = client.IndexHistory(context.Background(), indexReq)

	if err != nil {
		if vnstockErr, ok := err.(*vnstock.Error); ok {
			if vnstockErr.Code == vnstock.InvalidInput {
				log.Fatal("❌ Index date validation failed - should accept valid date range")
			}
		}
	}

	fmt.Println("✅ Index date validation works")
}
