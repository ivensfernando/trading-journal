package connectors

import (
	"context"
	kucoin "github.com/Kucoin/kucoin-go-sdk"
)

type KucoinConnector struct {
	apiService *kucoin.ApiService
}

func NewKucoinConnector(apiKey, apiSecret, apiPassphrase string) *KucoinConnector {
	apiService := kucoin.NewApiService(
		kucoin.ApiKeyOption(apiKey),
		kucoin.ApiSecretOption(apiSecret),
		kucoin.ApiPassPhraseOption(apiPassphrase),
	)

	return &KucoinConnector{apiService: apiService}
}

func (kc *KucoinConnector) TestConnection() error {
	ctx := context.Background()
	_, err := kc.apiService.ServerTime(ctx)
	return err
}

//func (kc *KucoinConnector) GetOrderBook(symbol string, limit int) (*kucoin.OrderBookModel, error) {
//	ctx := context.Background()
//	params := map[string]string{
//		"symbol": symbol,
//		"limit":  fmt.Sprintf("%d", limit),
//	}
//	rsp, err := kc.apiService.MarketOrderBook(ctx, params)
//	if err != nil {
//		return nil, err
//	}
//
//	var orderBook kucoin.OrderBookModel
//	if err := rsp.ReadData(&orderBook); err != nil {
//		return nil, err
//	}
//
//	return &orderBook, nil
//}

func (kc *KucoinConnector) ExecuteOrder(orderType, symbol string, quantity, price float64) (string, error) {
	// Implementação para execução de ordens
	return "", nil
}
