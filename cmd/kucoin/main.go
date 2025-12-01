package main

import (
	"fmt"
	"log"
	"os"

	ccxt "github.com/ccxt/ccxt/go/v4"
	"vsC1Y2025V01/internal/connectors"
)

func main() {
	apiKey := os.Getenv("KUCOIN_API_KEY")
	apiSecret := os.Getenv("KUCOIN_API_SECRET")
	passphrase := os.Getenv("KUCOIN_API_PASSPHRASE")

	if apiKey == "" || apiSecret == "" || passphrase == "" {
		log.Fatal("environment variables KUCOIN_API_KEY, KUCOIN_API_SECRET and KUCOIN_API_PASSPHRASE are required")
	}

	connector := connectors.NewKucoinConnector(ccxt.Credentials{
		ApiKey:     apiKey,
		Secret:     apiSecret,
		Passphrase: passphrase,
	})

	if err := connector.TestConnection(); err != nil {
		log.Fatalf("failed to reach KuCoin: %v", err)
	}

	balances, err := connector.GetAccountBalances()
	if err != nil {
		log.Fatalf("could not fetch balances: %v", err)
	}

	fmt.Println("== KuCoin balances ==")
	for currency, amount := range balances {
		fmt.Printf("%-12s %f\n", currency, amount)
	}
}
