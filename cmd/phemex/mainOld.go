package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	// Testnet REST base URL
	baseURL = "https://testnet-api.phemex.com"
)

// Standard Phemex response wrapper
type APIResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// Structure for USDT-M Futures returned by /g-accounts/positions
type GAccountPositions struct {
	Account struct {
		UserID           int64  `json:"userID"`
		AccountID        int64  `json:"accountId"`
		Currency         string `json:"currency"`
		AccountBalanceRv string `json:"accountBalanceRv"` // actual USDT balance
	} `json:"account"`

	Positions []struct {
		AccountID        int64  `json:"accountID"`
		Symbol           string `json:"symbol"`
		Currency         string `json:"currency"`
		Side             string `json:"side"`    // Buy / Sell, but for orders only
		PosSide          string `json:"posSide"` // Long / Short / Merged
		SizeRq           string `json:"sizeRq"`  // size in contract units (string)
		AvgEntryPriceRp  string `json:"avgEntryPriceRp"`
		PositionMarginRv string `json:"positionMarginRv"`
		MarkPriceRp      string `json:"markPriceRp"`
	} `json:"positions"`
}

// ---------------------------------
// AUTHENTICATED CLIENT
// ---------------------------------

type Client struct {
	apiKey    string
	apiSecret string
	http      *http.Client
}

// Create client
func NewClient(apiKey, apiSecret string) *Client {
	return &Client{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		http:      &http.Client{Timeout: 10 * time.Second},
	}
}

// Sign request using Phemex formula:
// signature = HMAC_SHA256(path + query + expiry + body)
func signRequest(path, query, body string, expiry int64, apiSecret string) string {
	signString := path
	if query != "" {
		signString += query
	}
	signString += fmt.Sprintf("%d", expiry)
	if body != "" {
		signString += body
	}

	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write([]byte(signString))
	return hex.EncodeToString(mac.Sum(nil))
}

// Generic request handler
func (c *Client) doRequest(method, path, query string, body []byte) (*APIResponse, error) {

	expiry := time.Now().Add(1 * time.Minute).Unix()

	bodyStr := ""
	if body != nil {
		bodyStr = string(body)
	}

	signature := signRequest(path, query, bodyStr, expiry, c.apiSecret)

	url := baseURL + path
	if query != "" {
		url += "?" + query
	}

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-phemex-access-token", c.apiKey)
	req.Header.Set("x-phemex-request-expiry", fmt.Sprintf("%d", expiry))
	req.Header.Set("x-phemex-request-signature", signature)

	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(rawBody))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(rawBody, &apiResp); err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	return &apiResp, nil
}

// ---------------------------------
// B) LIST USDT FUTURES POSITIONS
// ---------------------------------

func (c *Client) ListUSDTPositions() error {
	path := "/g-accounts/positions"
	query := "currency=USDT"

	fmt.Println("→ Fetching USDT futures positions...")

	resp, err := c.doRequest(http.MethodGet, path, query, nil)
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("API error: code=%d msg=%s", resp.Code, resp.Msg)
	}

	var ap GAccountPositions
	if err := json.Unmarshal(resp.Data, &ap); err != nil {
		return fmt.Errorf("failed parsing positions: %w", err)
	}

	fmt.Printf("USDT Balance: %s\n", ap.Account.AccountBalanceRv)

	found := false

	for _, p := range ap.Positions {
		if p.SizeRq == "" || p.SizeRq == "0" {
			continue
		}

		found = true
		fmt.Println("------ OPEN POSITION ------")
		fmt.Printf("Symbol:     %s\n", p.Symbol)
		fmt.Printf("PosSide:    %s\n", p.PosSide)
		fmt.Printf("SizeRq:     %s\n", p.SizeRq)
		fmt.Printf("AvgPrice:   %s\n", p.AvgEntryPriceRp)
		fmt.Printf("Margin:     %s\n", p.PositionMarginRv)
		fmt.Printf("MarkPrice:  %s\n", p.MarkPriceRp)
		fmt.Println("---------------------------")
	}

	if !found {
		fmt.Println("No open USDT-M positions.")
	}

	return nil
}

// ---------------------------------
// C1) OPEN ETHUSDT LONG (MARKET)
// ---------------------------------

func (c *Client) OpenEthUsdtLong(orderQty string) error {
	path := "/g-orders"

	body := map[string]interface{}{
		"symbol":      "ETHUSDT",
		"clOrdID":     fmt.Sprintf("ivens-open-%d", time.Now().UnixNano()),
		"side":        "Buy",    // order direction
		"posSide":     "Long",   // REQUIRED for hedge mode
		"ordType":     "Market", // Market order
		"orderQtyRq":  orderQty, // e.g. "0.01"
		"reduceOnly":  false,    // open / increase pos
		"timeInForce": "ImmediateOrCancel",
	}

	bodyBytes, _ := json.Marshal(body)

	fmt.Println("→ Opening LONG ETHUSDT (Market)...")
	resp, err := c.doRequest(http.MethodPost, path, "", bodyBytes)
	if err != nil {
		return err
	}

	fmt.Printf("code=%d msg=%s\n", resp.Code, resp.Msg)
	fmt.Printf("data=%s\n", string(resp.Data))
	return nil
}

// ---------------------------------
// C2) CLOSE ETHUSDT LONG (MARKET)
// ---------------------------------

func (c *Client) CloseEthUsdtLong(orderQty string) error {
	path := "/g-orders"

	body := map[string]interface{}{
		"symbol":      "ETHUSDT",
		"clOrdID":     fmt.Sprintf("ivens-close-%d", time.Now().UnixNano()),
		"side":        "Sell", // opposite direction
		"posSide":     "Long", // close LONG leg
		"ordType":     "Market",
		"orderQtyRq":  orderQty,
		"reduceOnly":  true, // IMPORTANT to prevent accidental increase
		"timeInForce": "ImmediateOrCancel",
	}

	bodyBytes, _ := json.Marshal(body)

	fmt.Println("→ Closing LONG ETHUSDT (Market, reduceOnly)...")
	resp, err := c.doRequest(http.MethodPost, path, "", bodyBytes)
	if err != nil {
		return err
	}

	fmt.Printf("code=%d msg=%s\n", resp.Code, resp.Msg)
	fmt.Printf("data=%s\n", string(resp.Data))
	return nil
}

// ---------------------------------
// MAIN EXECUTION
// ---------------------------------

func mains() {

	apiKey := os.Getenv("PHEMEX_API_KEY")
	apiSecret := os.Getenv("PHEMEX_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		log.Fatal("Set PHEMEX_API_KEY and PHEMEX_API_SECRET env variables.")
	}

	client := NewClient(apiKey, apiSecret)

	// List USDT futures positions
	if err := client.ListUSDTPositions(); err != nil {
		log.Fatalf("ListUSDTPositions error: %v", err)
	}

	// Example: open 0.01 ETH long
	_ = client.OpenEthUsdtLong("0.01")

	//Example: close 0.01 ETH long?
	//_ = client.CloseEthUsdtLong("0.01")
}
