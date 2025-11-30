package kucoinuniversal

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultSpotBaseURL    = "https://api.kucoin.com"
	defaultFuturesBaseURL = "https://api-futures.kucoin.com"
)

// Config configures the KuCoin universal client.
type Config struct {
	APIKey         string
	APISecret      string
	APIPassphrase  string
	SpotBaseURL    string
	FuturesBaseURL string
	HTTPClient     *http.Client
	KeyVersion     string
	// EncryptPassphrase controls whether the KC-API-PASSPHRASE header is HMAC + base64
	// encoded with the API secret. Some key setups (typically legacy key version 2)
	// require the encrypted form, while others expect the plain passphrase. Default
	// behaviour keeps encryption enabled for backwards compatibility.
	EncryptPassphrase bool
}

// Client provides access to the KuCoin REST endpoints used by the connector.
type Client struct {
	apiKey         string
	apiSecret      string
	apiPassphrase  string
	spotBaseURL    string
	futuresBaseURL string
	keyVersion     string
	httpClient     *http.Client
	encryptPass    bool
}

// NewClient builds a Client using the provided configuration.
func NewClient(cfg Config) (*Client, error) {
	if cfg.APIKey == "" || cfg.APISecret == "" {
		return nil, errors.New("api key and secret are required")
	}

	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	spotURL := cfg.SpotBaseURL
	if spotURL == "" {
		spotURL = defaultSpotBaseURL
	}

	futuresURL := cfg.FuturesBaseURL
	if futuresURL == "" {
		futuresURL = defaultFuturesBaseURL
	}

	keyVersion := cfg.KeyVersion
	if keyVersion == "" {
		keyVersion = "3"
	}

	if cfg.APIPassphrase == "" && keyVersion != "3" {
		return nil, errors.New("api passphrase is required for key version 2")
	}

	return &Client{
		apiKey:         cfg.APIKey,
		apiSecret:      cfg.APISecret,
		apiPassphrase:  cfg.APIPassphrase,
		spotBaseURL:    strings.TrimSuffix(spotURL, "/"),
		futuresBaseURL: strings.TrimSuffix(futuresURL, "/"),
		keyVersion:     keyVersion,
		httpClient:     client,
		encryptPass:    cfg.EncryptPassphrase || cfg.KeyVersion == "2" || cfg.KeyVersion == "",
	}, nil
}

// ServerTime fetches the KuCoin server time in milliseconds.
func (c *Client) ServerTime(ctx context.Context) (time.Time, error) {
	payload, err := c.doRequest(ctx, http.MethodGet, c.spotBaseURL, "/api/v1/timestamp", nil, nil, false)
	if err != nil {
		return time.Time{}, err
	}

	var rsp struct {
		Code string `json:"code"`
		Data int64  `json:"data"`
	}

	if err := json.Unmarshal(payload, &rsp); err != nil {
		return time.Time{}, fmt.Errorf("failed to decode server time: %w", err)
	}

	return time.UnixMilli(rsp.Data), nil
}

// UserInfo returns account level information from the v2 endpoint.
type UserInfo struct {
	UID             string `json:"uid"`
	UserLevel       string `json:"userLevel"`
	SubName         string `json:"subName"`
	Type            string `json:"type"`
	TradeEnabled    bool   `json:"tradeEnabled"`
	TransferEnabled bool   `json:"transferEnabled"`
}

// GetAccountSummaryInfo fetches account summary details using the v2 endpoint.
func (c *Client) GetAccountSummaryInfo(ctx context.Context) (*UserInfo, error) {
	payload, err := c.doRequest(ctx, http.MethodGet, c.spotBaseURL, "/api/v2/user-info", nil, nil, true)
	if err != nil {
		return nil, err
	}

	var rsp struct {
		Code string    `json:"code"`
		Data *UserInfo `json:"data"`
	}

	if err := json.Unmarshal(payload, &rsp); err != nil {
		return nil, fmt.Errorf("failed to decode account summary response: %w", err)
	}

	return rsp.Data, nil
}

// AccountBalance mirrors the response for spot accounts.
type AccountBalance struct {
	ID        string `json:"id"`
	Currency  string `json:"currency"`
	Type      string `json:"type"`
	Balance   string `json:"balance"`
	Available string `json:"available"`
	Holds     string `json:"holds"`
}

// IsolatedMarginAccount represents a simplified isolated margin account entry.
type IsolatedMarginAccount struct {
	Symbol    string          `json:"symbol"`
	Assets    json.RawMessage `json:"assets"`
	Liability json.RawMessage `json:"liability"`
	RiskRate  string          `json:"riskRate"`
}

// GetIsolatedMarginAccounts returns isolated margin accounts (v3 endpoint).
func (c *Client) GetIsolatedMarginAccounts(ctx context.Context) ([]IsolatedMarginAccount, error) {
	payload, err := c.doRequest(ctx, http.MethodGet, c.spotBaseURL, "/api/v3/isolated/accounts", nil, nil, true)
	if err != nil {
		return nil, err
	}

	var rsp struct {
		Code string                  `json:"code"`
		Data []IsolatedMarginAccount `json:"data"`
	}

	if err := json.Unmarshal(payload, &rsp); err != nil {
		return nil, fmt.Errorf("failed to decode isolated accounts: %w", err)
	}

	return rsp.Data, nil
}

// DepositAddress defines a deposit address entry.
type DepositAddress struct {
	Currency        string `json:"currency"`
	Chain           string `json:"chain"`
	Address         string `json:"address"`
	Memo            string `json:"memo"`
	ContractAddress string `json:"contractAddress"`
}

// GetDepositAddresses fetches deposit addresses for a currency (v3 endpoint).
func (c *Client) GetDepositAddresses(ctx context.Context, currency, chain string) ([]DepositAddress, error) {
	query := url.Values{}
	if currency != "" {
		query.Set("currency", currency)
	}
	if chain != "" {
		query.Set("chain", chain)
	}

	payload, err := c.doRequest(ctx, http.MethodGet, c.spotBaseURL, "/api/v3/deposit-addresses", query, nil, true)
	if err != nil {
		return nil, err
	}

	var rsp struct {
		Code string           `json:"code"`
		Data []DepositAddress `json:"data"`
	}

	if err := json.Unmarshal(payload, &rsp); err != nil {
		return nil, fmt.Errorf("failed to decode deposit addresses: %w", err)
	}

	return rsp.Data, nil
}

// WithdrawalRequest wraps the payload for a withdrawal (v3 endpoint).
type WithdrawalRequest struct {
	Currency        string  `json:"currency"`
	Amount          float64 `json:"amount"`
	Address         string  `json:"address"`
	Chain           string  `json:"chain,omitempty"`
	Memo            string  `json:"memo,omitempty"`
	Remark          string  `json:"remark,omitempty"`
	FeeDeductType   string  `json:"feeDeductType,omitempty"`
	IsInnerTransfer bool    `json:"isInnerTransfer,omitempty"`
}

// WithdrawalResponse captures the identifier returned by the withdrawal API.
type WithdrawalResponse struct {
	WithdrawalID string `json:"withdrawalId"`
}

// CreateWithdrawal submits a withdrawal using the v3 endpoint.
func (c *Client) CreateWithdrawal(ctx context.Context, req WithdrawalRequest) (*WithdrawalResponse, error) {
	payload, err := c.doRequest(ctx, http.MethodPost, c.spotBaseURL, "/api/v3/withdrawals", nil, req, true)
	if err != nil {
		return nil, err
	}

	var rsp struct {
		Code string             `json:"code"`
		Data WithdrawalResponse `json:"data"`
	}

	if err := json.Unmarshal(payload, &rsp); err != nil {
		return nil, fmt.Errorf("failed to decode withdrawal response: %w", err)
	}

	return &rsp.Data, nil
}

// FlexTransferRequest represents universal transfer payload.
type FlexTransferRequest struct {
	ClientOID string  `json:"clientOid"`
	From      string  `json:"from"`
	To        string  `json:"to"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
}

// FlexTransferResponse captures the transferId returned by KuCoin.
type FlexTransferResponse struct {
	TransferID string `json:"transferId"`
}

// FlexTransfer performs an internal flexible transfer (universal transfer).
func (c *Client) FlexTransfer(ctx context.Context, req FlexTransferRequest) (*FlexTransferResponse, error) {
	payload, err := c.doRequest(ctx, http.MethodPost, c.spotBaseURL, "/api/v3/flex-transfer", nil, req, true)
	if err != nil {
		return nil, err
	}

	var rsp struct {
		Code string               `json:"code"`
		Data FlexTransferResponse `json:"data"`
	}

	if err := json.Unmarshal(payload, &rsp); err != nil {
		return nil, fmt.Errorf("failed to decode transfer response: %w", err)
	}

	return &rsp.Data, nil
}

// PurchaseOrder wraps data returned by purchase order queries.
type PurchaseOrder struct {
	OrderID string `json:"orderId"`
	Symbol  string `json:"symbol"`
	Status  string `json:"status"`
	Side    string `json:"side"`
}

// ListPurchaseOrders queries credit purchase orders (v3 endpoint).
func (c *Client) ListPurchaseOrders(ctx context.Context, status string) ([]PurchaseOrder, error) {
	query := url.Values{}
	if status != "" {
		query.Set("status", status)
	}

	payload, err := c.doRequest(ctx, http.MethodGet, c.spotBaseURL, "/api/v3/purchase/orders", query, nil, true)
	if err != nil {
		return nil, err
	}

	var rsp struct {
		Code string          `json:"code"`
		Data []PurchaseOrder `json:"data"`
	}

	if err := json.Unmarshal(payload, &rsp); err != nil {
		return nil, fmt.Errorf("failed to decode purchase orders: %w", err)
	}

	return rsp.Data, nil
}

// HighFrequencyLedgerEntry models a margin HF ledger row.
type HighFrequencyLedgerEntry struct {
	LedgerID string  `json:"ledgerId"`
	Amount   float64 `json:"amount"`
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
	BizType  string  `json:"bizType"`
	Created  int64   `json:"createdAt"`
}

// GetMarginHFLedgers fetches margin high-frequency account ledgers.
func (c *Client) GetMarginHFLedgers(ctx context.Context, currency string) ([]HighFrequencyLedgerEntry, error) {
	query := url.Values{}
	if currency != "" {
		query.Set("currency", currency)
	}

	payload, err := c.doRequest(ctx, http.MethodGet, c.spotBaseURL, "/api/v3/account-ledgers-marginhf", query, nil, true)
	if err != nil {
		return nil, err
	}

	var rsp struct {
		Code string                     `json:"code"`
		Data []HighFrequencyLedgerEntry `json:"data"`
	}

	if err := json.Unmarshal(payload, &rsp); err != nil {
		return nil, fmt.Errorf("failed to decode margin HF ledgers: %w", err)
	}

	return rsp.Data, nil
}

// SubAPIKey captures a sub-account API key entry.
type SubAPIKey struct {
	SubName    string `json:"subName"`
	APIKey     string `json:"apiKey"`
	Passphrase string `json:"passPhrase"`
	Permission string `json:"permission"`
}

// ListSubAPIKeys lists sub-account API keys (v1 endpoint).
func (c *Client) ListSubAPIKeys(ctx context.Context, subName string) ([]SubAPIKey, error) {
	query := url.Values{}
	if subName != "" {
		query.Set("subName", subName)
	}

	payload, err := c.doRequest(ctx, http.MethodGet, c.spotBaseURL, "/api/v1/sub/api-key", query, nil, true)
	if err != nil {
		return nil, err
	}

	var rsp struct {
		Code string      `json:"code"`
		Data []SubAPIKey `json:"data"`
	}

	if err := json.Unmarshal(payload, &rsp); err != nil {
		return nil, fmt.Errorf("failed to decode sub api keys: %w", err)
	}

	return rsp.Data, nil
}

// GetSpotAccounts fetches spot account balances.
func (c *Client) GetSpotAccounts(ctx context.Context) ([]AccountBalance, error) {
	payload, err := c.doRequest(ctx, http.MethodGet, c.spotBaseURL, "/api/v1/accounts", nil, nil, true)
	if err != nil {
		return nil, err
	}

	var rsp struct {
		Code string           `json:"code"`
		Data []AccountBalance `json:"data"`
	}

	if err := json.Unmarshal(payload, &rsp); err != nil {
		return nil, fmt.Errorf("failed to decode account response: %w", err)
	}

	return rsp.Data, nil
}

// SpotOrderRequest represents the payload for placing spot orders.
type SpotOrderRequest struct {
	ClientOID string  `json:"clientOid,omitempty"`
	Symbol    string  `json:"symbol"`
	Side      string  `json:"side"`
	Type      string  `json:"type"`
	Size      float64 `json:"size,omitempty"`
	Funds     float64 `json:"funds,omitempty"`
	Price     float64 `json:"price,omitempty"`
	Stop      string  `json:"stop,omitempty"`
	StopPrice float64 `json:"stopPrice,omitempty"`
}

// SpotOrderResponse captures the minimal response fields for spot orders.
type SpotOrderResponse struct {
	OrderID string `json:"orderId"`
	Symbol  string `json:"symbol"`
	Side    string `json:"side"`
	Type    string `json:"type"`
	Price   string `json:"price"`
	Size    string `json:"size"`
	Funds   string `json:"funds"`
	Status  string `json:"status"`
}

// CreateSpotOrder submits a new spot order.
func (c *Client) CreateSpotOrder(ctx context.Context, req SpotOrderRequest) (*SpotOrderResponse, error) {
	payload, err := c.doRequest(ctx, http.MethodPost, c.spotBaseURL, "/api/v1/orders", nil, req, true)
	if err != nil {
		return nil, err
	}

	var rsp struct {
		Code string            `json:"code"`
		Data SpotOrderResponse `json:"data"`
	}

	if err := json.Unmarshal(payload, &rsp); err != nil {
		return nil, fmt.Errorf("failed to decode spot order response: %w", err)
	}

	return &rsp.Data, nil
}

// GetSpotOrder retrieves an existing spot order by ID.
func (c *Client) GetSpotOrder(ctx context.Context, orderID string) (*SpotOrderResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/orders/%s", url.PathEscape(orderID))
	payload, err := c.doRequest(ctx, http.MethodGet, c.spotBaseURL, endpoint, nil, nil, true)
	if err != nil {
		return nil, err
	}

	var rsp struct {
		Code string            `json:"code"`
		Data SpotOrderResponse `json:"data"`
	}

	if err := json.Unmarshal(payload, &rsp); err != nil {
		return nil, fmt.Errorf("failed to decode spot order query: %w", err)
	}

	return &rsp.Data, nil
}

// FuturesOrderRequest represents the payload for futures orders.
type FuturesOrderRequest struct {
	ClientOID       string  `json:"clientOid,omitempty"`
	Symbol          string  `json:"symbol"`
	Side            string  `json:"side"`
	Type            string  `json:"type"`
	Leverage        float64 `json:"leverage,omitempty"`
	Size            float64 `json:"size"`
	Price           float64 `json:"price,omitempty"`
	StopLossPrice   float64 `json:"stopLossPrice,omitempty"`
	TakeProfitPrice float64 `json:"takeProfitPrice,omitempty"`
}

// FuturesOrderResponse captures minimal futures order details.
type FuturesOrderResponse struct {
	OrderID string `json:"orderId"`
	Symbol  string `json:"symbol"`
	Side    string `json:"side"`
	Type    string `json:"type"`
	Price   string `json:"price"`
	Size    string `json:"size"`
	Status  string `json:"status"`
}

// CreateFuturesOrder submits a futures order.
func (c *Client) CreateFuturesOrder(ctx context.Context, req FuturesOrderRequest) (*FuturesOrderResponse, error) {
	payload, err := c.doRequest(ctx, http.MethodPost, c.futuresBaseURL, "/api/v1/orders", nil, req, true)
	if err != nil {
		return nil, err
	}

	var rsp struct {
		Code string               `json:"code"`
		Data FuturesOrderResponse `json:"data"`
	}

	if err := json.Unmarshal(payload, &rsp); err != nil {
		return nil, fmt.Errorf("failed to decode futures order response: %w", err)
	}

	return &rsp.Data, nil
}

// UpdateFuturesStops adjusts stop loss and take profit for an order.
type UpdateFuturesStopsRequest struct {
	StopLossPrice   float64 `json:"stopLossPrice,omitempty"`
	TakeProfitPrice float64 `json:"takeProfitPrice,omitempty"`
}

// LeverageUpdateRequest updates user leverage on futures or margin positions.
type LeverageUpdateRequest struct {
	Symbol     string  `json:"symbol"`
	Leverage   float64 `json:"leverage"`
	MarginMode string  `json:"marginMode"`
}

// LeverageUpdateResponse mirrors the minimal response from the leverage endpoint.
type LeverageUpdateResponse struct {
	Symbol   string  `json:"symbol"`
	Leverage float64 `json:"leverage"`
	Mode     string  `json:"marginMode"`
}

// UpdateUserLeverage modifies leverage using the v3 endpoint.
func (c *Client) UpdateUserLeverage(ctx context.Context, req LeverageUpdateRequest) (*LeverageUpdateResponse, error) {
	payload, err := c.doRequest(ctx, http.MethodPost, c.spotBaseURL, "/api/v3/position/update-user-leverage", nil, req, true)
	if err != nil {
		return nil, err
	}

	var rsp struct {
		Code string                 `json:"code"`
		Data LeverageUpdateResponse `json:"data"`
	}

	if err := json.Unmarshal(payload, &rsp); err != nil {
		return nil, fmt.Errorf("failed to decode leverage update response: %w", err)
	}

	return &rsp.Data, nil
}

// UpdateFuturesStops updates stop-loss and take-profit prices for a futures order.
func (c *Client) UpdateFuturesStops(ctx context.Context, orderID string, req UpdateFuturesStopsRequest) (*FuturesOrderResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/orders/%s", url.PathEscape(orderID))
	payload, err := c.doRequest(ctx, http.MethodPut, c.futuresBaseURL, endpoint, nil, req, true)
	if err != nil {
		return nil, err
	}

	var rsp struct {
		Code string               `json:"code"`
		Data FuturesOrderResponse `json:"data"`
	}

	if err := json.Unmarshal(payload, &rsp); err != nil {
		return nil, fmt.Errorf("failed to decode futures stop update response: %w", err)
	}

	return &rsp.Data, nil
}

// GetFuturesOrder retrieves an existing futures order by ID.
func (c *Client) GetFuturesOrder(ctx context.Context, orderID string) (*FuturesOrderResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/orders/%s", url.PathEscape(orderID))
	payload, err := c.doRequest(ctx, http.MethodGet, c.futuresBaseURL, endpoint, nil, nil, true)
	if err != nil {
		return nil, err
	}

	var rsp struct {
		Code string               `json:"code"`
		Data FuturesOrderResponse `json:"data"`
	}

	if err := json.Unmarshal(payload, &rsp); err != nil {
		return nil, fmt.Errorf("failed to decode futures order query: %w", err)
	}

	return &rsp.Data, nil
}

func (c *Client) doRequest(ctx context.Context, method, baseURL, path string, query url.Values, body interface{}, signed bool) ([]byte, error) {
	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	endpoint := path
	if len(query) > 0 {
		encoded := query.Encode()
		if strings.Contains(endpoint, "?") {
			endpoint += "&" + encoded
		} else {
			endpoint += "?" + encoded
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if signed {
		c.signRequest(req, endpoint, payload)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("kucoin api error %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	return raw, nil
}

func (c *Client) signRequest(req *http.Request, endpoint string, body []byte) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signPayload := timestamp + req.Method + endpoint
	if len(body) > 0 {
		signPayload += string(body)
	}

	mac := hmac.New(sha256.New, []byte(c.apiSecret))
	mac.Write([]byte(signPayload))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	passphrase := c.apiPassphrase
	if passphrase != "" && c.encryptPass {
		mac := hmac.New(sha256.New, []byte(c.apiSecret))
		mac.Write([]byte(c.apiPassphrase))
		passphrase = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	}

	req.Header.Set("KC-API-SIGN", signature)
	req.Header.Set("KC-API-TIMESTAMP", timestamp)
	req.Header.Set("KC-API-KEY", c.apiKey)
	if passphrase != "" {
		req.Header.Set("KC-API-PASSPHRASE", passphrase)
	}
	req.Header.Set("KC-API-KEY-VERSION", c.keyVersion)
}
