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

type MexcCredentials struct {
	ApiKey string
	Secret string
}

type MexcClient struct {
	credentials MexcCredentials
	baseURL     string
	httpClient  *http.Client
}

const (
	mexcSpotBaseURL    = "https://api.mexc.com"
	mexcFuturesBaseURL = "https://contract.mexc.com"
)

func NewMexcSpotClient(credentials MexcCredentials) *MexcClient {
	return &MexcClient{
		credentials: credentials,
		baseURL:     mexcSpotBaseURL,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}
}

func NewMexcFuturesClient(credentials MexcCredentials) *MexcClient {
	return &MexcClient{
		credentials: credentials,
		baseURL:     mexcFuturesBaseURL,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (m *MexcClient) Ping(ctx context.Context) error {
	path := "/api/v3/ping"
	if m.baseURL == mexcFuturesBaseURL {
		path = "/api/v1/ping"
	}

	_, err := m.doRequest(ctx, http.MethodGet, path, nil, false)
	return err
}

type MexcSpotBalance struct {
	Asset  string  `json:"asset"`
	Free   float64 `json:"free"`
	Locked float64 `json:"locked"`
}

type MexcFuturesBalance struct {
	Currency         string  `json:"currency"`
	AvailableBalance float64 `json:"availableBalance"`
	Equity           float64 `json:"equity"`
}

func (m *MexcClient) FetchSpotBalances(ctx context.Context) ([]MexcSpotBalance, error) {
	params := url.Values{}
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

	body, err := m.doRequest(ctx, http.MethodGet, "/api/v3/account?"+params.Encode(), nil, true)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Balances []struct {
			Asset  string `json:"asset"`
			Free   string `json:"free"`
			Locked string `json:"locked"`
		} `json:"balances"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse spot balances: %w", err)
	}

	balances := make([]MexcSpotBalance, 0, len(resp.Balances))
	for _, bal := range resp.Balances {
		free, err := strconv.ParseFloat(bal.Free, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid free amount for %s: %w", bal.Asset, err)
		}
		locked, err := strconv.ParseFloat(bal.Locked, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid locked amount for %s: %w", bal.Asset, err)
		}

		balances = append(balances, MexcSpotBalance{
			Asset:  bal.Asset,
			Free:   free,
			Locked: locked,
		})
	}

	return balances, nil
}

func (m *MexcClient) FetchFuturesBalance(ctx context.Context, currency string) (*MexcFuturesBalance, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("currency", currency)
	}
	params.Set("req_time", strconv.FormatInt(time.Now().UnixMilli(), 10))

	path := "/api/v1/private/account/asset"
	if query := params.Encode(); query != "" {
		path = path + "?" + query
	}

	body, err := m.doRequest(ctx, http.MethodGet, path, nil, true)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data struct {
			Currency         string `json:"currency"`
			AvailableBalance string `json:"availableBalance"`
			Equity           string `json:"equity"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse futures balance: %w", err)
	}

	available, err := strconv.ParseFloat(resp.Data.AvailableBalance, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid futures available balance: %w", err)
	}
	equity, err := strconv.ParseFloat(resp.Data.Equity, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid futures equity balance: %w", err)
	}

	return &MexcFuturesBalance{
		Currency:         resp.Data.Currency,
		AvailableBalance: available,
		Equity:           equity,
	}, nil
}

func (m *MexcClient) doRequest(ctx context.Context, method, path string, body []byte, signed bool) ([]byte, error) {
	endpoint := m.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	if signed {
		query := req.URL.RawQuery
		if query == "" && strings.Contains(path, "?") {
			query = strings.SplitN(path, "?", 2)[1]
		}
		signature := m.sign(query)
		if query != "" {
			req.URL.RawQuery = query + "&signature=" + signature
		} else {
			req.URL.RawQuery = "signature=" + signature
		}
		req.Header.Set("X-MEXC-APIKEY", m.credentials.ApiKey)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return io.ReadAll(resp.Body)
}

func (m *MexcClient) sign(payload string) string {
	mac := hmac.New(sha256.New, []byte(m.credentials.Secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
