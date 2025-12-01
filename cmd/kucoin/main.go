package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	ccxt "github.com/ccxt/ccxt/go/v4"
	"vsC1Y2025V01/internal/connectors"
)

func main() {
	apiKey := os.Getenv("KUCOIN_API_KEY")
	apiSecret := os.Getenv("KUCOIN_API_SECRET")
	passphrase := os.Getenv("KUCOIN_API_PASSPHRASE")
	keyVersion := os.Getenv("KUCOIN_API_KEY_VERSION")

	if apiKey == "" || apiSecret == "" || passphrase == "" {
		log.Fatal("environment variables KUCOIN_API_KEY, KUCOIN_API_SECRET and KUCOIN_API_PASSPHRASE are required")
	}

	credentials := ccxt.Credentials{
		ApiKey:     apiKey,
		Secret:     apiSecret,
		Passphrase: passphrase,
	}

	if keyVersion != "" {
		version, err := strconv.Atoi(keyVersion)
		if err != nil {
			log.Fatalf("invalid KUCOIN_API_KEY_VERSION: %v", err)
		}
		credentials.KeyVersion = version
	}

	connector := connectors.NewKucoinConnector(credentials)

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
