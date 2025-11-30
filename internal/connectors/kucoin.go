package connectors

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	kucoin "github.com/Kucoin/kucoin-universal-sdk"
)

type KucoinConnector struct {
	client *kucoin.Client
}

func NewKucoinConnector(apiKey, apiSecret, apiPassphrase, keyVersion string) (*KucoinConnector, error) {
	client, err := kucoin.NewClient(kucoin.Config{
		APIKey:        apiKey,
		APISecret:     apiSecret,
		APIPassphrase: apiPassphrase,
		KeyVersion:    keyVersion,
	})
	if err != nil {
		return nil, err
	}

	return &KucoinConnector{client: client}, nil
}

func (kc *KucoinConnector) TestConnection() error {
	ctx := context.Background()
	_, err := kc.client.ServerTime(ctx)
	return err
}

func (kc *KucoinConnector) GetAccountBalances() (map[string]float64, error) {
	ctx := context.Background()
	accounts, err := kc.client.GetSpotAccounts(ctx)
	if err != nil {
		return nil, err
	}

	balances := make(map[string]float64)
	for _, account := range accounts {
		available, err := strconv.ParseFloat(account.Available, 64)
		if err != nil {
			continue
		}
		balances[account.Currency] += available
	}

	return balances, nil
}

func (kc *KucoinConnector) ExecuteOrder(orderType, symbol string, quantity, price float64) (string, error) {
	ctx := context.Background()
	side := "buy"
	if quantity < 0 {
		side = "sell"
		quantity = -quantity
	}

	switch strings.ToLower(orderType) {
	case "spot", "spot-limit", "spot-market":
		orderTypeName := "limit"
		if strings.Contains(orderType, "market") {
			orderTypeName = "market"
		}

		order, err := kc.client.CreateSpotOrder(ctx, kucoin.SpotOrderRequest{
			Symbol: symbol,
			Side:   side,
			Type:   orderTypeName,
			Size:   quantity,
			Price:  price,
		})
		if err != nil {
			return "", err
		}
		return order.OrderID, nil
	case "futures", "futures-limit", "futures-market":
		orderTypeName := "limit"
		if strings.Contains(orderType, "market") {
			orderTypeName = "market"
		}

		order, err := kc.client.CreateFuturesOrder(ctx, kucoin.FuturesOrderRequest{
			Symbol: symbol,
			Side:   side,
			Type:   orderTypeName,
			Size:   quantity,
			Price:  price,
		})
		if err != nil {
			return "", err
		}
		return order.OrderID, nil
	default:
		return "", fmt.Errorf("unsupported order type: %s", orderType)
	}
}

func (kc *KucoinConnector) GetSpotOrder(orderID string) (*kucoin.SpotOrderResponse, error) {
	return kc.client.GetSpotOrder(context.Background(), orderID)
}

func (kc *KucoinConnector) GetFuturesOrder(orderID string) (*kucoin.FuturesOrderResponse, error) {
	return kc.client.GetFuturesOrder(context.Background(), orderID)
}

func (kc *KucoinConnector) UpdateFuturesStops(orderID string, stopLoss, takeProfit float64) (*kucoin.FuturesOrderResponse, error) {
	return kc.client.UpdateFuturesStops(context.Background(), orderID, kucoin.UpdateFuturesStopsRequest{
		StopLossPrice:   stopLoss,
		TakeProfitPrice: takeProfit,
	})
}
