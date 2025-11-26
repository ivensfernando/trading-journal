package payloads

type AlertPayload struct {
	Ticker    string `json:"ticker validate:required"`
	Action    string `json:"action validate:required"`
	Sentiment string `json:"sentiment"`
	Quantity  string `json:"quantity validate:required"`
	Price     string `json:"price validate:required"`
	Time      string `json:"time"`
	Interval  string `json:"interval"`
}

type CreateWebhookPayload struct {
	Name        string `json:"name validate:required"`
	Type        string `json:"type validate:required"` //crypto stocks futures
	Description string `json:"description"`
	Style       string `json:"style"`
	Tickers     string `json:"tickers validate:required"`
}
