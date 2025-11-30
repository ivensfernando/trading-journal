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
}

// NewClient builds a Client using the provided configuration.
func NewClient(cfg Config) (*Client, error) {
	if cfg.APIKey == "" || cfg.APISecret == "" || cfg.APIPassphrase == "" {
		return nil, errors.New("api key, secret, and passphrase are required")
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
		keyVersion = "2"
	}

	return &Client{
		apiKey:         cfg.APIKey,
		apiSecret:      cfg.APISecret,
		apiPassphrase:  cfg.APIPassphrase,
		spotBaseURL:    strings.TrimSuffix(spotURL, "/"),
		futuresBaseURL: strings.TrimSuffix(futuresURL, "/"),
		keyVersion:     keyVersion,
		httpClient:     client,
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

// AccountBalance mirrors the response for spot accounts.
type AccountBalance struct {
	ID        string `json:"id"`
	Currency  string `json:"currency"`
	Type      string `json:"type"`
	Balance   string `json:"balance"`
	Available string `json:"available"`
	Holds     string `json:"holds"`
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

	mac.Reset()
	mac.Write([]byte(c.apiPassphrase))
	passphrase := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req.Header.Set("KC-API-SIGN", signature)
	req.Header.Set("KC-API-TIMESTAMP", timestamp)
	req.Header.Set("KC-API-KEY", c.apiKey)
	req.Header.Set("KC-API-PASSPHRASE", passphrase)
	req.Header.Set("KC-API-KEY-VERSION", c.keyVersion)
}
