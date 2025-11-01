package main

import (
	"context"
	"fmt"
	"github.com/Kucoin/kucoin-go-sdk"
	"log"
	"os"
)

func main() {
	// Fetch KuCoin credentials from environment variables
	apiKey := os.Getenv("KUCOIN_API_KEY")
	apiSecret := os.Getenv("KUCOIN_API_SECRET")
	apiPassphrase := os.Getenv("KUCOIN_API_PASSPHRASE")

	if apiKey == "" || apiSecret == "" || apiPassphrase == "" {
		log.Fatal("KuCoin credentials are not set in the environment variables.")
	}

	// Initialize KuCoin API client
	//apiService := connectors.NewKucoinConnector(apiKey, apiSecret, apiPassphrase)

	client := kucoin.NewApiService(
		kucoin.ApiKeyOption(apiKey),
		kucoin.ApiSecretOption(apiSecret),
		kucoin.ApiPassPhraseOption(apiPassphrase),
	)

	// Test connection to the KuCoin API
	serverTime, err := client.ServerTime(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch server time: %v", err)
	}
	fmt.Printf("Connected to KuCoin. Server time: %v\n", serverTime)
	//
	//// Fetch account balances
	//fetchAccountBalances(client)
}

func fetchAccountBalances(client *kucoin.ApiService) {
	// Request account balances
	//resp, err := client.Accounts()
	//if err != nil {
	//	log.Fatalf("Failed to fetch account balances: %v", err)
	//}
	//
	//// Print account balances
	//for _, account := range resp.Data {
	//	fmt.Printf("Currency: %s, Available: %s, Balance: %s\n", account.Currency, account.Available, account.Balance)
	//}
}
