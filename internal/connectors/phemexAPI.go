// phemex_api_client.go
// FULL REST API CLIENT FOR PHEMEX USDT-M FUTURES
// Includes all sections A → E
// A) Authenticated Client
// B) Account & Positions
// C) Trading Methods
// D) Order Query Methods
// E) Market Data Methods
// Logging levels: INFO, DEBUG, ERROR

package connectors

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
	"time"
)

// -----------------------------
// CONFIG
// -----------------------------
const (
	baseURL = "https://testnet-api.phemex.com" // Change to prod if needed
)

// -----------------------------
// LOGGING HELPERS
// -----------------------------
func logInfo(msg string, args ...interface{})  { log.Printf("[INFO] "+msg, args...) }
func logDebug(msg string, args ...interface{}) { log.Printf("[DEBUG] "+msg, args...) }
func logError(msg string, args ...interface{}) { log.Printf("[ERROR] "+msg, args...) }

// -----------------------------
// API RESPONSE WRAPPER
// -----------------------------
type APIResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// -----------------------------
// B) STRUCTURES FOR POSITIONS
// -----------------------------
type GAccountPositions struct {
	Account struct {
		UserID           int64  `json:"userID"`
		AccountID        int64  `json:"accountId"`
		Currency         string `json:"currency"`
		AccountBalanceRv string `json:"accountBalanceRv"`
	} `json:"account"`

	Positions []struct {
		AccountID        int64  `json:"accountID"`
		Symbol           string `json:"symbol"`
		Currency         string `json:"currency"`
		Side             string `json:"side"`
		PosSide          string `json:"posSide"`
		SizeRq           string `json:"sizeRq"`
		AvgEntryPriceRp  string `json:"avgEntryPriceRp"`
		PositionMarginRv string `json:"positionMarginRv"`
		MarkPriceRp      string `json:"markPriceRp"`
	} `json:"positions"`
}

// -----------------------------
// A) AUTHENTICATED CLIENT
// -----------------------------
type Client struct {
	apiKey    string
	apiSecret string
	http      *http.Client
}

func NewClient(apiKey, apiSecret string) *Client {
	return &Client{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		http:      &http.Client{Timeout: 15 * time.Second},
	}
}

func signRequest(path, query, body string, expiry int64, secret string) string {
	base := path
	if query != "" {
		base += query
	}
	base += fmt.Sprintf("%d", expiry)
	if body != "" {
		base += body
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(base))
	return hex.EncodeToString(mac.Sum(nil))
}

func (c *Client) doRequest(method, path, query string, body []byte) (*APIResponse, error) {
	expiry := time.Now().Add(1 * time.Minute).Unix()

	bodyStr := ""
	if body != nil {
		bodyStr = string(body)
	}

	sig := signRequest(path, query, bodyStr, expiry, c.apiSecret)

	url := baseURL + path
	if query != "" {
		url += "?" + query
	}

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	logDebug("HTTP %s %s", method, url)

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-phemex-access-token", c.apiKey)
	req.Header.Set("x-phemex-request-expiry", fmt.Sprintf("%d", expiry))
	req.Header.Set("x-phemex-request-signature", sig)

	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		logError("Request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	logDebug("Response: %s", string(raw))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(raw))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(raw, &apiResp); err != nil {
		return nil, err
	}

	return &apiResp, nil
}

// -----------------------------
// B) ACCOUNT & POSITION METHODS
// -----------------------------
func (c *Client) GetPositionsUSDT() (*GAccountPositions, error) {
	logInfo("Fetching USDT positions...")

	resp, err := c.doRequest("GET", "/g-accounts/positions", "currency=USDT", nil)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: %s", resp.Msg)
	}

	var parsed GAccountPositions
	if err := json.Unmarshal(resp.Data, &parsed); err != nil {
		return nil, err
	}

	return &parsed, nil
}

// -----------------------------
// C) TRADING METHODS
// -----------------------------
func (c *Client) PlaceOrder(symbol, side, posSide, qty, ordType string, reduce bool) (*APIResponse, error) {
	body := map[string]interface{}{
		"symbol":      symbol,
		"side":        side,
		"posSide":     posSide,
		"ordType":     ordType,
		"orderQtyRq":  qty,
		"reduceOnly":  reduce,
		"clOrdID":     fmt.Sprintf("go-%d", time.Now().UnixNano()),
		"timeInForce": "ImmediateOrCancel",
	}

	b, _ := json.Marshal(body)
	logInfo("Placing order %s %s qty=%s", side, symbol, qty)

	return c.doRequest("POST", "/g-orders", "", b)
}

func (c *Client) CancelAll(symbol string) (*APIResponse, error) {
	logInfo("Canceling all orders for %s", symbol)
	return c.doRequest("DELETE", "/g-orders/all", fmt.Sprintf("symbol=%s", symbol), nil)
}

// -----------------------------
// D) ORDER QUERY METHODS
// -----------------------------
func (c *Client) GetActiveOrders(symbol string) (*APIResponse, error) {
	logInfo("Fetching active orders for %s", symbol)
	return c.doRequest("GET", "/g-orders/activeList", fmt.Sprintf("symbol=%s", symbol), nil)
}

func (c *Client) GetOrderHistory(symbol string) (*APIResponse, error) {
	logInfo("Fetching order history for %s", symbol)
	return c.doRequest("GET", "/g-orders/history", fmt.Sprintf("symbol=%s", symbol), nil)
}

func (c *Client) GetFills(symbol string) (*APIResponse, error) {
	logInfo("Fetching fills for %s", symbol)
	return c.doRequest("GET", "/g-trades/fills", fmt.Sprintf("symbol=%s", symbol), nil)
}

// -----------------------------
// E) MARKET DATA METHODS
// -----------------------------
func (c *Client) GetTicker(symbol string) (*APIResponse, error) {
	return c.doRequest("GET", "/md/v3/ticker/24hr", fmt.Sprintf("symbol=%s", symbol), nil)
}

func (c *Client) GetOrderbook(symbol string) (*APIResponse, error) {
	return c.doRequest("GET", "/md/v2/orderbook", fmt.Sprintf("symbol=%s", symbol), nil)
}

func (c *Client) GetKlines(symbol string, res int) (*APIResponse, error) {
	return c.doRequest("GET", "/md/perpetual/kline", fmt.Sprintf("symbol=%s&resolution=%d", symbol, res), nil)
}
