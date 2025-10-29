package connectors

import (
	"context"
	"log"

	mexcMarketData "github.com/linstohu/nexapi/mexc/spot/marketdata"
	mexcTypes "github.com/linstohu/nexapi/mexc/spot/marketdata/types"
	mexcUtils "github.com/linstohu/nexapi/mexc/spot/utils"
)

type MexcConnector struct {
	marketDataClient *mexcMarketData.SpotMarketDataClient
}

func NewMexcConnector(apiKey, apiSecret string) *MexcConnector {
	cfg := &mexcUtils.SpotClientCfg{
		Debug:      true,
		BaseURL:    mexcUtils.BaseURL,
		Key:        apiKey,
		Secret:     apiSecret,
		RecvWindow: 5000,
	}

	marketDataClient, err := mexcMarketData.NewSpotMarketDataClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize MEXC market data client: %v", err)
	}

	return &MexcConnector{marketDataClient: marketDataClient}
}

func (mc *MexcConnector) TestConnection() error {
	err := mc.marketDataClient.Ping(context.Background())
	return err
}

func (mc *MexcConnector) GetOrderBook(symbol string, limit int) (*mexcTypes.Orderbook, error) {
	params := mexcTypes.GetOrderbookParams{
		Symbol: symbol,
		Limit:  limit,
	}

	return mc.marketDataClient.GetOrderbook(context.Background(), params)

}

func (mc *MexcConnector) ExecuteOrder(orderType, symbol string, quantity, price float64) (string, error) {
	// Placeholder for order execution implementation
	return "", nil
}
