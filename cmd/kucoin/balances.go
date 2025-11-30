//go:build kucoinbalances
// +build kucoinbalances

package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	kucoin "github.com/Kucoin/kucoin-universal-sdk"
)

func main() {
	key := os.Getenv("KUCOIN_API_KEY")
	secret := os.Getenv("KUCOIN_API_SECRET")
	passphrase := os.Getenv("KUCOIN_API_PASSPHRASE")

	keyVersion := strings.TrimSpace(os.Getenv("KUCOIN_API_KEY_VERSION"))
	encryptPassphrase := parseBoolEnv("KUCOIN_ENCRYPT_PASSPHRASE", true)

	client, err := kucoin.NewClient(kucoin.Config{
		APIKey:            key,
		APISecret:         secret,
		APIPassphrase:     passphrase,
		KeyVersion:        keyVersion,
		EncryptPassphrase: encryptPassphrase,
	})
	if err != nil {
		log.Fatalf("failed to build KuCoin client: %v", err)
	}

	ctx := context.Background()

	spotBalances, err := client.GetSpotAccounts(ctx)
	if err != nil {
		log.Fatalf("failed to fetch spot balances: %v", err)
	}
	log.Println("Spot wallets (available):")
	for currency, available := range aggregateAvailable(spotBalances) {
		log.Printf("- %s: %.8f", currency, available)
	}

	currency := os.Getenv("KUCOIN_FUTURES_CURRENCY")
	if currency == "" {
		currency = "USDT"
	}

	futuresBalances, err := client.GetFuturesAccountOverview(ctx, currency)
	if err != nil {
		log.Fatalf("failed to fetch futures balance: %v", err)
	}
	log.Printf("Futures wallet: available=%s %s, equity=%s", futuresBalances.AvailableBalance, futuresBalances.Currency, futuresBalances.AccountEquity)

	marginAccount, err := client.GetMarginAccount(ctx)
	if err != nil {
		log.Fatalf("failed to fetch margin balances: %v", err)
	}
	log.Println("Cross margin wallets (available):")
	for _, account := range marginAccount.Accounts {
		available, err := strconv.ParseFloat(account.AvailableBalance, 64)
		if err != nil {
			log.Printf("- %s: could not parse available balance (%s): %v", account.Currency, account.AvailableBalance, err)
			continue
		}
		log.Printf("- %s: %.8f (total=%s, liability=%s)", account.Currency, available, account.TotalBalance, account.Liability)
	}
}

func aggregateAvailable(accounts []kucoin.AccountBalance) map[string]float64 {
	balances := make(map[string]float64)
	for _, account := range accounts {
		available, err := strconv.ParseFloat(account.Available, 64)
		if err != nil {
			continue
		}
		balances[account.Currency] += available
	}
	return balances
}

func parseBoolEnv(name string, defaultVal bool) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(name)))
	switch raw {
	case "", "unset":
		return defaultVal
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return defaultVal
	}
}
