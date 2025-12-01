package connectors

import (
	"context"
	"errors"
	"fmt"
	"log"

	ccxt "github.com/ccxt/ccxt/go/v4"
)

// MexcConnector implements ExchangeConnector using spot and futures CCXT clients.
type MexcConnector struct {
	spotClient    *ccxt.MexcClient
	futuresClient *ccxt.MexcClient
}

// NewMexcConnector builds a connector configured with the provided credentials.
func NewMexcConnector(apiKey, secret string) *MexcConnector {
	credentials := ccxt.MexcCredentials{ApiKey: apiKey, Secret: secret}
	return &MexcConnector{
		spotClient:    ccxt.NewMexcSpotClient(credentials),
		futuresClient: ccxt.NewMexcFuturesClient(credentials),
	}
}

// TestConnection checks connectivity against both spot and futures APIs.
func (m *MexcConnector) TestConnection() error {
	ctx := context.Background()

	log.Printf("pinging MEXC spot API at %s", m.spotClient.BaseURL())

	if err := m.spotClient.Ping(ctx); err != nil {
		return fmt.Errorf("spot ping failed: %w", err)
	}

	log.Printf("pinging MEXC futures API at %s", m.futuresClient.BaseURL())

	if err := m.futuresClient.Ping(ctx); err != nil {
		return fmt.Errorf("futures ping failed: %w", err)
	}

	return nil
}

// GetAccountBalances aggregates balances from the spot and futures accounts.
func (m *MexcConnector) GetAccountBalances() (map[string]float64, error) {
	ctx := context.Background()
	balances := make(map[string]float64)

	spotBalances, err := m.spotClient.FetchSpotBalances(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch spot balances: %w", err)
	}

	for _, bal := range spotBalances {
		if bal.Free == 0 && bal.Locked == 0 {
			continue
		}
		balances[fmt.Sprintf("spot_%s", bal.Asset)] = bal.Free
	}

	futuresBalance, err := m.futuresClient.FetchFuturesBalance(ctx, "USDT")
	if err != nil {
		return nil, fmt.Errorf("fetch futures balance: %w", err)
	}

	if futuresBalance != nil {
		balances[fmt.Sprintf("futures_%s", futuresBalance.Currency)] = futuresBalance.AvailableBalance
	}

	return balances, nil
}

// ExecuteOrder is not implemented for this example connector.
func (m *MexcConnector) ExecuteOrder(orderType string, symbol string, quantity float64, price float64) (string, error) {
	return "", errors.New("order execution is not implemented in the sample MEXC connector")
}

var _ ExchangeConnector = (*MexcConnector)(nil)
