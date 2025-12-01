package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sirupsen/logrus"

	"vsC1Y2025V01/internal/connectors"
)

func main() {
	apiKey := os.Getenv("PHEMEX_API_KEY")
	secret := os.Getenv("PHEMEX_API_SECRET")

	if apiKey == "" || secret == "" {
		log.Fatal("environment variables PHEMEX_API_KEY and PHEMEX_API_SECRET are required")
	}

	connector := connectors.NewPhemexConnector(apiKey, secret)

	if err := connector.TestConnection(); err != nil {
		log.Fatalf("failed to connect to Phemex: %v", err)
	}

	balances, err := connector.GetAccountBalances()
	if err != nil {
		log.Fatalf("failed to fetch Phemex balances: %v", err)
	}

	logrus.Infof("Balances: %v", balances)

	fmt.Println("Phemex connection successful and balances retrieved.")
}
