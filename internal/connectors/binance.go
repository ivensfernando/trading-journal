package connectors

import (
	"context"
	"encoding/json"
	"fmt"

	bnspotmd "github.com/linstohu/nexapi/binance/spot/marketdata"
	bnspottypes "github.com/linstohu/nexapi/binance/spot/marketdata/types"
	bnspotutils "github.com/linstohu/nexapi/binance/spot/utils"
)

func main() {
	cli, err := bnspotmd.NewSpotMarketDataClient(&bnspotutils.SpotClientCfg{
		Debug:   true,
		BaseURL: bnspotutils.BaseURL,
	})
	if err != nil {
		panic(err)
	}

	orderbook, err := cli.GetOrderbook(context.TODO(), bnspottypes.GetOrderbookParams{
		Symbol: "BTCUSDT",
		Limit:  5,
	})
	if err != nil {
		panic(err)
	}

	limit := orderbook.Http.ApiRes.Header.Get("X-Mbx-Used-Weight-1m")

	fmt.Printf("Current used request weight: %v\n", limit)

	bytes, err := json.MarshalIndent(orderbook.Body, "", "\t")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bytes))
}
