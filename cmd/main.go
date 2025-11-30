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

	if apiKey == "" || apiSecret == "" || apiPassphrase == "" {
		log.Fatal("KuCoin credentials are not set in the environment variables.")
	}

	client, err := kucoin.NewClient(kucoin.Config{
		APIKey:        apiKey,
		APISecret:     apiSecret,
		APIPassphrase: apiPassphrase,
	})
	if err != nil {
		log.Fatalf("Failed to initialize KuCoin client: %v", err)
	}

	serverTime, err := client.ServerTime(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch server time: %v", err)
	}
	fmt.Printf("Connected to KuCoin. Server time: %v\n", serverTime)
}
