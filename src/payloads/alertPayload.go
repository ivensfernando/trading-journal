package payloads

import "time"

type AlertPayload struct {
	Ticker                 string     `json:"ticker"`
	Action                 string     `json:"action"`
	Sentiment              string     `json:"sentiment"`
	Quantity               float64    `json:"quantity"`
	Price                  float64    `json:"price"`
	Time                   string     `json:"time"`
	Interval               string     `json:"interval"`
	MarketPosition         string     `json:"marketPosition"`
	PrevMarketPosition     string     `json:"prevMarketPosition"`
	MarketPositionSize     float64    `json:"marketPositionSize"`
	PrevMarketPositionSize float64    `json:"prevMarketPositionSize"`
	AlertTime              *time.Time `json:"alert_time"`
}
