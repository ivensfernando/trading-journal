// phemex_api_client.go
// REST API client for Phemex USDT-M Futures
// Sections: A) Authenticated Client, B) List Positions, C) Trading Methods
// Includes logging at multiple levels (INFO, DEBUG, ERROR)

package connectors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// -----------------------------------------
// B) LIST POSITIONS
// -----------------------------------------
func (c *Client) ListUSDTPositions() (*GAccountPositions, error) {
	path := "/g-accounts/positions"
	query := "currency=USDT"

	logInfo("Fetching USDT futures positions...")

	resp, err := c.doRequest(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("API error %d: %s", resp.Code, resp.Msg)
	}

	var positions GAccountPositions
	if err := json.Unmarshal(resp.Data, &positions); err != nil {
		return nil, err
	}

	return &positions, nil
}

// -----------------------------------------
// C) TRADING METHODS
// -----------------------------------------

// Open ETHUSDT LONG Market
func (c *Client) OpenEthLong(qty string) error {
	path := "/g-orders"
	body := map[string]interface{}{
		"symbol":      "ETHUSDT",
		"clOrdID":     fmt.Sprintf("go-long-%d", time.Now().UnixNano()),
		"side":        "Buy",
		"posSide":     "Long",
		"ordType":     "Market",
		"orderQtyRq":  qty,
		"reduceOnly":  false,
		"timeInForce": "ImmediateOrCancel",
	}

	b, _ := json.Marshal(body)
	logInfo("Opening LONG ETHUSDT with qty=%s", qty)

	resp, err := c.doRequest(http.MethodPost, path, "", b)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return fmt.Errorf("Order rejected: %s", resp.Msg)
	}

	logInfo("Order accepted: %s", resp.Msg)
	return nil
}

// Close LONG position
func (c *Client) CloseEthLong(qty string) error {
	path := "/g-orders"
	body := map[string]interface{}{
		"symbol":      "ETHUSDT",
		"clOrdID":     fmt.Sprintf("go-close-%d", time.Now().UnixNano()),
		"side":        "Sell",
		"posSide":     "Long",
		"ordType":     "Market",
		"orderQtyRq":  qty,
		"reduceOnly":  true,
		"timeInForce": "ImmediateOrCancel",
	}

	b, _ := json.Marshal(body)
	logInfo("Closing LONG ETHUSDT with qty=%s", qty)

	resp, err := c.doRequest(http.MethodPost, path, "", b)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return fmt.Errorf("Close rejected: %s", resp.Msg)
	}

	logInfo("Position closed: %s", resp.Msg)
	return nil
}

//// -----------------------------------------
//// MAIN
//// -----------------------------------------
//func main() {
//	apiKey := os.Getenv("PHEMEX_API_KEY")
//	apiSecret := os.Getenv("PHEMEX_API_SECRET")
//
//	if apiKey == "" || apiSecret == "" {
//		log.Fatal("Missing PHEMEX_API_KEY or PHEMEX_API_SECRET")
//	}
//
//	client := NewClient(apiKey, apiSecret)
//
//	// Example: list positions
//	pos, err := client.ListUSDTPositions()
//	if err != nil {
//		logError("Failed listing positions: %v", err)
//		return
//	}
//
//	logInfo("USDT Balance: %s", pos.Account.AccountBalanceRv)
//
//	// Uncomment to open/close orders
//	// client.OpenEthLong("0.01")
//	// client.CloseEthLong("0.01")
//}
