package connectors

import (
	"context"
	"log"

	mexcMarketData "github.com/linstohu/nexapi/mexc/spot/marketdata"
	mexcTypes "github.com/linstohu/nexapi/mexc/spot/marketdata/types"
	mexcUtils "github.com/linstohu/nexapi/mexc/spot/utils"
)

// MarketDataClient é uma interface para o cliente de dados de mercado
type MarketDataClient interface {
	Ping(ctx context.Context) error
	GetOrderbook(ctx context.Context, params mexcTypes.GetOrderbookParams) (*mexcTypes.Orderbook, error)
}

type MexcConnector struct {
	marketDataClient MarketDataClient
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
	return mc.marketDataClient.Ping(context.Background())
}

func (mc *MexcConnector) GetOrderBook(symbol string, limit int) (*mexcTypes.Orderbook, error) {
	params := mexcTypes.GetOrderbookParams{
		Symbol: symbol,
		Limit:  limit,
	}

	return mc.marketDataClient.GetOrderbook(context.Background(), params)
}

func (mc *MexcConnector) ExecuteOrder(orderType, symbol string, quantity, price float64) (string, error) {
	// Placeholder para implementação da execução de ordens
	return "", nil
}
