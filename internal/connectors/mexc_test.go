package connectors

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"

	mexcTypes "github.com/linstohu/nexapi/mexc/spot/marketdata/types"
	"github.com/stretchr/testify/mock"
)

// MockSpotMarketDataClient é um mock para SpotMarketDataClient
type MockSpotMarketDataClient struct {
	mock.Mock
}

func (m *MockSpotMarketDataClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSpotMarketDataClient) GetOrderbook(ctx context.Context, params mexcTypes.GetOrderbookParams) (*mexcTypes.Orderbook, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*mexcTypes.Orderbook), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestNewMexcConnector(t *testing.T) {
	apiKey := "test-key"
	apiSecret := "test-secret"

	connector := NewMexcConnector(apiKey, apiSecret)
	assert.NotNil(t, connector)
	assert.NotNil(t, connector.marketDataClient)
}

func TestMexcConnector_TestConnection(t *testing.T) {
	mockClient := new(MockSpotMarketDataClient)
	mockClient.On("Ping", mock.Anything).Return(nil)

	connector := &MexcConnector{marketDataClient: mockClient}

	err := connector.TestConnection()
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestMexcConnector_TestConnection_Fail(t *testing.T) {
	mockClient := new(MockSpotMarketDataClient)
	mockClient.On("Ping", mock.Anything).Return(errors.New("connection error"))

	connector := &MexcConnector{marketDataClient: mockClient}

	err := connector.TestConnection()
	assert.Error(t, err)
	assert.Equal(t, "connection error", err.Error())
	mockClient.AssertExpectations(t)
}

func TestMexcConnector_GetOrderBook(t *testing.T) {
	mockClient := new(MockSpotMarketDataClient)
	orderbook := &mexcTypes.Orderbook{
		LastUpdateID: 12345,
	}
	mockClient.On("GetOrderbook", mock.Anything, mock.Anything).Return(orderbook, nil)

	connector := &MexcConnector{marketDataClient: mockClient}

	result, err := connector.GetOrderBook("BTCUSDT", 5)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(12345), result.LastUpdateID)
	mockClient.AssertExpectations(t)
}

func TestMexcConnector_GetOrderBook_Fail(t *testing.T) {
	mockClient := new(MockSpotMarketDataClient)
	mockClient.On("GetOrderbook", mock.Anything, mock.Anything).Return(nil, errors.New("orderbook error"))

	connector := &MexcConnector{marketDataClient: mockClient}

	result, err := connector.GetOrderBook("BTCUSDT", 5)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "orderbook error", err.Error())
	mockClient.AssertExpectations(t)
}

func TestMexcConnector_ExecuteOrder(t *testing.T) {
	connector := &MexcConnector{}
	orderID, err := connector.ExecuteOrder("LIMIT", "BTCUSDT", 1.0, 50000.0)
	assert.NoError(t, err)
	assert.Equal(t, "", orderID) // Placeholder response
}
