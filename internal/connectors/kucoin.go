package connectors

import (
	"context"
	"errors"
	"fmt"

	ccxt "github.com/ccxt/ccxt/go/v4"
)

// KucoinConnector implements ExchangeConnector using the ccxt KuCoin client for spot and futures.
type KucoinConnector struct {
	spotClient    *ccxt.KucoinClient
	futuresClient *ccxt.KucoinClient
}

// NewKucoinConnector builds a connector configured with the provided credentials.
func NewKucoinConnector(credentials ccxt.Credentials) *KucoinConnector {
	return &KucoinConnector{
		spotClient:    ccxt.NewKucoinClient(credentials),
		futuresClient: ccxt.NewKucoinFuturesClient(credentials),
	}
}

// TestConnection checks connectivity against both spot and futures APIs.
func (k *KucoinConnector) TestConnection() error {
	ctx := context.Background()

	if err := k.spotClient.Ping(ctx); err != nil {
		return fmt.Errorf("spot ping failed: %w", err)
	}

	if err := k.futuresClient.Ping(ctx); err != nil {
		return fmt.Errorf("futures ping failed: %w", err)
	}

	return nil
}

// GetAccountBalances aggregates balances from the spot and futures accounts.
func (k *KucoinConnector) GetAccountBalances() (map[string]float64, error) {
	ctx := context.Background()
	balances := make(map[string]float64)

	spotBalances, err := k.spotClient.FetchSpotBalances(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch spot balances: %w", err)
	}

	for _, balance := range spotBalances {
		if balance.Available == 0 {
			continue
		}
		balances[fmt.Sprintf("spot_%s", balance.Currency)] = balance.Available
	}

	futuresBalance, err := k.futuresClient.FetchFuturesBalance(ctx, "USDT")
	if err != nil {
		return nil, fmt.Errorf("fetch futures balance: %w", err)
	}

	if futuresBalance != nil {
		balances[fmt.Sprintf("futures_%s", futuresBalance.Currency)] = futuresBalance.AvailableBalance
	}

	return balances, nil
}

// ExecuteOrder is not implemented for this example connector.
func (k *KucoinConnector) ExecuteOrder(orderType string, symbol string, quantity float64, price float64) (string, error) {
	return "", errors.New("order execution is not implemented in the sample KuCoin connector")
}
