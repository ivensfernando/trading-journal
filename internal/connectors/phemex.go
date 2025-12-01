package connectors

import (
	"context"
	"errors"
	"fmt"
	"log"

	ccxt "github.com/ccxt/ccxt/go/v4"
)

// PhemexConnector implements ExchangeConnector using the custom Phemex client.
type PhemexConnector struct {
	client *ccxt.PhemexClient
}

// NewPhemexConnector builds a connector configured with the provided credentials.
func NewPhemexConnector(apiKey, secret string) *PhemexConnector {
	return NewPhemexConnectorWithBaseURL(apiKey, secret, ccxt.DefaultPhemexBaseURL)
}

// NewPhemexConnectorWithBaseURL lets callers override the API host (e.g., testnet).
func NewPhemexConnectorWithBaseURL(apiKey, secret, baseURL string) *PhemexConnector {
	credentials := ccxt.PhemexCredentials{ApiKey: apiKey, Secret: secret}
	return &PhemexConnector{client: ccxt.NewPhemexClientWithBaseURL(credentials, baseURL)}
}

// TestConnection checks connectivity against the public ping endpoint.
func (p *PhemexConnector) TestConnection() error {
	ctx := context.Background()
	log.Printf("pinging Phemex API at %s", p.client.BaseURL())
	return p.client.Ping(ctx)
}

// GetAccountBalances fetches contract account balances and normalizes them to decimal amounts.
func (p *PhemexConnector) GetAccountBalances() (map[string]float64, error) {
	ctx := context.Background()
	balances := make(map[string]float64)

	contractBalances, err := p.client.FetchContractBalance(ctx, "USDT")
	if err != nil {
		return nil, fmt.Errorf("fetch contract balance: %w", err)
	}

	for _, bal := range contractBalances {
		// available balances are returned in Ev (1e-4) units for USDT
		available := float64(bal.AvailableBalanceEv) / 1e4
		balances[fmt.Sprintf("contract_%s", bal.Currency)] = available
	}

	return balances, nil
}

// ExecuteOrder is not implemented for this example connector.
func (p *PhemexConnector) ExecuteOrder(orderType string, symbol string, quantity float64, price float64) (string, error) {
	return "", errors.New("order execution is not implemented in the sample Phemex connector")
}

var _ ExchangeConnector = (*PhemexConnector)(nil)
