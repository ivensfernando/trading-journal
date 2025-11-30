package main

import (
	"context"
	"fmt"
	"log"
	"os"

	kucoin "github.com/Kucoin/kucoin-universal-sdk"
)

func main() {
	apiKey := os.Getenv("KUCOIN_API_KEY")
	apiSecret := os.Getenv("KUCOIN_API_SECRET")
	apiPassphrase := os.Getenv("KUCOIN_API_PASSPHRASE")
	apiKeyVersion := os.Getenv("KUCOIN_API_KEY_VERSION")
	if apiKeyVersion == "" {
		apiKeyVersion = "3"
	}

	if apiKey == "" || apiSecret == "" {
		log.Fatal("KuCoin credentials are not set in the environment variables.")
	}

	client, err := kucoin.NewClient(kucoin.Config{
		APIKey:        apiKey,
		APISecret:     apiSecret,
		APIPassphrase: apiPassphrase,
		KeyVersion:    apiKeyVersion,
	})
	if err != nil {
		log.Fatalf("Failed to initialize KuCoin client: %v", err)
	}

	ctx := context.Background()

	serverTime, err := client.ServerTime(ctx)
	if err != nil {
		log.Fatalf("Failed to fetch server time: %v", err)
	}
	fmt.Printf("Connected to KuCoin. Server time: %v\n", serverTime)

	accounts, err := client.GetSpotAccounts(ctx)
	if err != nil {
		log.Fatalf("Failed to fetch wallet balances: %v", err)
	}

	fmt.Println("Spot Wallet Balances:")
	for _, account := range accounts {
		fmt.Printf("- %s (%s): balance=%s, available=%s, holds=%s\n", account.Currency, account.Type, account.Balance, account.Available, account.Holds)
	}
}
