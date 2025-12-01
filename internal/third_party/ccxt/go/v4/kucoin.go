package ccxt

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Credentials holds the authentication data required by KuCoin.
type Credentials struct {
	ApiKey     string
	Secret     string
	Passphrase string
	KeyVersion int
}

// KucoinClient is a trimmed-down client inspired by the ccxt KuCoin connector.
type KucoinClient struct {
	credentials Credentials
	baseURL     string
	httpClient  *http.Client
}

const (
	spotBaseURL    = "https://api.kucoin.com"
	futuresBaseURL = "https://api-futures.kucoin.com"
)

// NewKucoinClient builds a client for the spot API.
func NewKucoinClient(credentials Credentials) *KucoinClient {
	return &KucoinClient{
		credentials: credentials,
		baseURL:     spotBaseURL,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}
}

// NewKucoinFuturesClient builds a client for the futures API.
func NewKucoinFuturesClient(credentials Credentials) *KucoinClient {
	return &KucoinClient{
		credentials: credentials,
		baseURL:     futuresBaseURL,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}
}

// Ping checks whether the exchange API is reachable.
func (k *KucoinClient) Ping(ctx context.Context) error {
	_, err := k.doRequest(ctx, http.MethodGet, "/api/v1/timestamp", nil, false)
	return err
}

// SpotBalance represents a single currency balance on the spot account.
type SpotBalance struct {
	Currency  string  `json:"currency"`
	Balance   float64 `json:"balance"`
	Available float64 `json:"available"`
	Holds     float64 `json:"holds"`
}

// FuturesBalance holds the futures account summary for a specific currency.
type FuturesBalance struct {
	Currency         string  `json:"currency"`
	AccountEquity    float64 `json:"accountEquity"`
	AvailableBalance float64 `json:"availableBalance"`
}

// FetchSpotBalances returns the available balances for the spot account.
func (k *KucoinClient) FetchSpotBalances(ctx context.Context) ([]SpotBalance, error) {
	body, err := k.doRequest(ctx, http.MethodGet, "/api/v1/accounts", nil, true)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Code string `json:"code"`
		Data []struct {
			Currency  string `json:"currency"`
			Type      string `json:"type"`
			Balance   string `json:"balance"`
			Available string `json:"available"`
			Holds     string `json:"holds"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse spot balances: %w", err)
	}

	balances := make([]SpotBalance, 0, len(resp.Data))
	for _, row := range resp.Data {
		if row.Currency == "" {
			continue
		}

		balance, err := parseFloat(row.Balance)
		if err != nil {
			return nil, fmt.Errorf("invalid balance for %s: %w", row.Currency, err)
		}
		available, err := parseFloat(row.Available)
		if err != nil {
			return nil, fmt.Errorf("invalid available balance for %s: %w", row.Currency, err)
		}
		holds, err := parseFloat(row.Holds)
		if err != nil {
			return nil, fmt.Errorf("invalid hold balance for %s: %w", row.Currency, err)
		}

		balances = append(balances, SpotBalance{
			Currency:  row.Currency,
			Balance:   balance,
			Available: available,
			Holds:     holds,
		})
	}

	return balances, nil
}

// FetchFuturesBalance returns the futures account overview for the given currency.
func (k *KucoinClient) FetchFuturesBalance(ctx context.Context, currency string) (*FuturesBalance, error) {
	params := url.Values{}
	if currency != "" {
		params.Add("currency", currency)
	}

	path := "/api/v1/account-overview"
	if query := params.Encode(); query != "" {
		path = path + "?" + query
	}

	body, err := k.doRequest(ctx, http.MethodGet, path, nil, true)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Code string `json:"code"`
		Data struct {
			Currency         string `json:"currency"`
			AccountEquity    string `json:"accountEquity"`
			AvailableBalance string `json:"availableBalance"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse futures balance: %w", err)
	}

	equity, err := parseFloat(resp.Data.AccountEquity)
	if err != nil {
		return nil, fmt.Errorf("invalid futures equity: %w", err)
	}
	available, err := parseFloat(resp.Data.AvailableBalance)
	if err != nil {
		return nil, fmt.Errorf("invalid futures available balance: %w", err)
	}

	return &FuturesBalance{
		Currency:         resp.Data.Currency,
		AccountEquity:    equity,
		AvailableBalance: available,
	}, nil
}

func parseFloat(value string) (float64, error) {
	if value == "" {
		return 0, nil
	}

	return strconv.ParseFloat(value, 64)
}

func (k *KucoinClient) doRequest(ctx context.Context, method, path string, body []byte, signed bool) ([]byte, error) {
	endpoint := k.baseURL + path
	var payload io.Reader
	if len(body) > 0 {
		payload = bytes.NewBuffer(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("User-Agent", "ccxt-go-sample")
	req.Header.Set("Content-Type", "application/json")

	if signed {
		if err := k.applyAuth(method, path, body, req.Header); err != nil {
			return nil, err
		}
	}

	res, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("exchange error: status %d, body %s", res.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

func (k *KucoinClient) applyAuth(method, path string, body []byte, headers http.Header) error {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	prehash := timestamp + method + path
	if len(body) > 0 {
		prehash += string(body)
	}

	signature, err := k.sign(prehash)
	if err != nil {
		return err
	}

	passphrase := k.credentials.Passphrase
	if k.credentials.KeyVersion == 0 || k.credentials.KeyVersion == 2 {
		var err error
		passphrase, err = k.sign(k.credentials.Passphrase)
		if err != nil {
			return err
		}
	}

	headers.Set("KC-API-KEY", k.credentials.ApiKey)
	headers.Set("KC-API-SIGN", signature)
	headers.Set("KC-API-TIMESTAMP", timestamp)
	headers.Set("KC-API-PASSPHRASE", passphrase)
	if k.credentials.KeyVersion != 1 {
		headers.Set("KC-API-KEY-VERSION", "2")
	} else {
		headers.Set("KC-API-KEY-VERSION", "1")
	}

	return nil
}

func (k *KucoinClient) sign(payload string) (string, error) {
	mac := hmac.New(sha256.New, []byte(k.credentials.Secret))
	if _, err := mac.Write([]byte(payload)); err != nil {
		return "", fmt.Errorf("sign payload: %w", err)
	}

	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}
