package ccxt

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// PhemexCredentials holds API access details for authenticated requests.
type PhemexCredentials struct {
	ApiKey string
	Secret string
}

// PhemexClient is a lightweight client for a subset of the Phemex REST API.
type PhemexClient struct {
	credentials PhemexCredentials
	baseURL     string
	httpClient  *http.Client
}

const (
	// defaultPhemexUserAgent identifies the lightweight ccxt-based client without
	// tying it to the hosting application so requests look like a generic SDK
	// consumer instead of a specific project build.
	defaultPhemexUserAgent = "ccxt-go-phemex/1.0"
)

const (
	// DefaultPhemexBaseURL points to the production REST API.
	DefaultPhemexBaseURL = "https://api.phemex.com"
	// TestnetPhemexBaseURL points to the Phemex testnet REST API.
	TestnetPhemexBaseURL = "https://testnet-api.phemex.com"
)

// NewPhemexClient builds a client configured with the provided credentials.
func NewPhemexClient(credentials PhemexCredentials) *PhemexClient {
	return NewPhemexClientWithBaseURL(credentials, DefaultPhemexBaseURL)
}

// NewPhemexClientWithBaseURL builds a client that targets a custom base URL (e.g., testnet).
func NewPhemexClientWithBaseURL(credentials PhemexCredentials, baseURL string) *PhemexClient {
	if baseURL == "" {
		baseURL = DefaultPhemexBaseURL
	}

	return &PhemexClient{
		credentials: credentials,
		baseURL:     baseURL,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}
}

// BaseURL returns the configured REST API host.
func (p *PhemexClient) BaseURL() string {
	return p.baseURL
}

// Ping checks connectivity using the public ping endpoint.
func (p *PhemexClient) Ping(ctx context.Context) error {
	_, err := p.doRequest(ctx, http.MethodGet, "/exchange/public/ping", "", false)
	return err
}

// PhemexAccountBalance represents the available balance from the contract account response.
type PhemexAccountBalance struct {
	Currency            string `json:"currency"`
	AvailableBalanceEv  int64  `json:"availableBalanceEv"`
	AccountBalanceEv    int64  `json:"accountBalanceEv"`
	TotalUsedBalanceEv  int64  `json:"totalUsedBalanceEv"`
	TotalPositionMargin int64  `json:"totalPositionMarginEv"`
	Raw                 json.RawMessage
}

// FetchContractBalance retrieves contract account balances for the requested currency.
func (p *PhemexClient) FetchContractBalance(ctx context.Context, currency string) ([]PhemexAccountBalance, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("currency", currency)
	}

	path := "/accounts/accountPositions"
	if query := params.Encode(); query != "" {
		path = path + "?" + query
	}

	body, err := p.doRequest(ctx, http.MethodGet, path, "", true)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Account []struct {
				Currency           string          `json:"currency"`
				AvailableBalanceEv int64           `json:"availableBalanceEv"`
				AccountBalanceEv   int64           `json:"accountBalanceEv"`
				TotalUsedBalanceEv int64           `json:"totalUsedBalanceEv"`
				TotalPositionEv    int64           `json:"totalPositionMarginEv"`
				Raw                json.RawMessage `json:"-"`
			} `json:"account"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse contract balance: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("phemex error %d: %s", resp.Code, resp.Msg)
	}

	balances := make([]PhemexAccountBalance, 0, len(resp.Data.Account))
	for _, bal := range resp.Data.Account {
		balances = append(balances, PhemexAccountBalance{
			Currency:            bal.Currency,
			AvailableBalanceEv:  bal.AvailableBalanceEv,
			AccountBalanceEv:    bal.AccountBalanceEv,
			TotalUsedBalanceEv:  bal.TotalUsedBalanceEv,
			TotalPositionMargin: bal.TotalPositionEv,
			Raw:                 bal.Raw,
		})
	}

	return balances, nil
}

func (p *PhemexClient) doRequest(ctx context.Context, method, path, body string, signed bool) ([]byte, error) {
	var payload io.Reader
	if body != "" {
		payload = strings.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, p.baseURL+path, payload)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", defaultPhemexUserAgent)

	if signed {
		if p.credentials.ApiKey == "" || p.credentials.Secret == "" {
			return nil, fmt.Errorf("missing API credentials for signed request")
		}

		expires := time.Now().Add(60 * time.Second).Unix()
		signaturePath := strings.Replace(path, "?", "", 1)
		signaturePayload := signaturePath + strconv.FormatInt(expires, 10) + body
		signature := p.sign(signaturePayload)

		req.Header.Set("x-phemex-access-token", p.credentials.ApiKey)
		req.Header.Set("x-phemex-request-signature", signature)
		req.Header.Set("x-phemex-request-expiry", strconv.FormatInt(expires, 10))
	}

	if body != "" || signed {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s %s failed: %w", method, p.baseURL+path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	return io.ReadAll(resp.Body)
}

func (p *PhemexClient) sign(payload string) string {
	mac := hmac.New(sha256.New, []byte(p.credentials.Secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
