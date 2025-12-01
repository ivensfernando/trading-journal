package main

import (
	"fmt"
	"log"
	"os"
	"vsC1Y2025V01/internal/connectors"
)

func main() {
	apiKey := os.Getenv("MEXC_API_KEY")
	apiSecret := os.Getenv("MEXC_API_SECRET")

	log.Println(apiKey)
	log.Println(apiSecret)

	if apiKey == "" || apiSecret == "" {
		log.Fatal("environment variables MEXC_API_KEY and MEXC_API_SECRET are required")
	}

	connector := connectors.NewMexcConnector(apiKey, apiSecret)

	if err := connector.TestConnection(); err != nil {
		log.Fatalf("failed to reach MEXC: %v", err)
	}

	balances, err := connector.GetAccountBalances()
	if err != nil {
		log.Fatalf("could not fetch balances: %v", err)
	}

	fmt.Println("== MEXC balances ==")
	for currency, amount := range balances {
		fmt.Printf("%-12s %f\n", currency, amount)
	}
}
