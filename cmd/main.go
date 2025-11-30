package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	kucoin "github.com/Kucoin/kucoin-universal-sdk"
	"log"
	"os"
)

func main() {
	apiKey := os.Getenv("KUCOIN_API_KEY")
	apiSecret := os.Getenv("KUCOIN_API_SECRET")
	apiPassphrase := os.Getenv("KUCOIN_API_PASSPHRASE")
	apiKeyVersion := os.Getenv("KUCOIN_API_KEY_VERSION")
	if apiKeyVersion == "" {
		apiKeyVersion = "3"
	}

	//encryptPassphrase := true
	//if v := os.Getenv("KUCOIN_API_PASSPHRASE_ENCRYPTED"); v != "" {
	//	parsed, err := strconv.ParseBool(v)
	//	if err != nil {
	//		log.Fatalf("Invalid KUCOIN_API_PASSPHRASE_ENCRYPTED value: %v", err)
	//	}
	//	encryptPassphrase = parsed
	//}

	apiP := KucoinSignPassphrase(apiSecret, apiPassphrase)

	if apiKey == "" || apiSecret == "" {
		log.Fatal("KuCoin credentials are not set in the environment variables.")
	}

	client, err := kucoin.NewClient(kucoin.Config{
		APIKey:        apiKey,
		APISecret:     apiSecret,
		APIPassphrase: apiP,
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

	if summary, err := client.GetAccountSummaryInfo(ctx); err == nil && summary != nil {
		fmt.Printf("User %s (level %s) trade enabled: %t, transfer enabled: %t\n", summary.UID, summary.UserLevel, summary.TradeEnabled, summary.TransferEnabled)
	}

	accounts, err := client.GetSpotAccounts(ctx)
	if err != nil {
		log.Fatalf("Failed to fetch wallet balances: %v", err)
	}

	fmt.Println("Spot Wallet Balances:")
	for _, account := range accounts {
		fmt.Printf("- %s (%s): balance=%s, available=%s, holds=%s\n", account.Currency, account.Type, account.Balance, account.Available, account.Holds)
	}

	if currency := os.Getenv("KUCOIN_DEPOSIT_CURRENCY"); currency != "" {
		addresses, err := client.GetDepositAddresses(ctx, currency, os.Getenv("KUCOIN_DEPOSIT_CHAIN"))
		if err != nil {
			log.Fatalf("Failed to fetch deposit addresses: %v", err)
		}
		fmt.Printf("Deposit addresses for %s:\n", currency)
		for _, addr := range addresses {
			fmt.Printf("- %s on %s (memo: %s)\n", addr.Address, addr.Chain, addr.Memo)
		}
	}
}

func EncryptPassphrase(secret, passphrase string) (string, error) {
	h := hmac.New(sha256.New, []byte(secret))
	if _, err := h.Write([]byte(passphrase)); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

func KucoinSignPassphrase(passphrase, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(passphrase))
	signature := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(signature)
}
