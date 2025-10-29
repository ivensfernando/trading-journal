package connectors

// ExchangeConnector defines the interface for exchange connectors
type ExchangeConnector interface {
	TestConnection() error
	GetAccountBalances() (map[string]float64, error)
	ExecuteOrder(orderType string, symbol string, quantity float64, price float64) (string, error)
}
