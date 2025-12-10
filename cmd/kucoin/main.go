package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	logger "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"

	"vsC1Y2025V01/src/connectors"
)

func SetupLogger() {
	levelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))

	level, err := logger.ParseLevel(levelStr)
	if err != nil {
		level = logger.DebugLevel // fallback seguro
	}

	logger.SetLevel(level)
	logger.SetFormatter(&logger.TextFormatter{
		FullTimestamp: true,
	})
}

func printUsage() {
	fmt.Println("Available commands:")
	fmt.Println("  help                             Show this help message")
	fmt.Println("  shutdown                         Exit the application")
	fmt.Println("  positions                        List all USDT-M positions")
	fmt.Println("  long SYMBOL QTY                  Open LONG market position")
	fmt.Println("  short SYMBOL QTY                 Open SHORT market position")
	fmt.Println("  close-long SYMBOL QTY            Close LONG")
	fmt.Println("  close-short SYMBOL QTY           Close SHORT")
	fmt.Println("  reverse SYMBOL QTY               Reverse position")
	fmt.Println("  long-usdt SYMBOL USDT LEV        Long using USDT amount and leverage")
	fmt.Println("  short-usdt SYMBOL USDT LEV       Short using USDT amount and leverage")
	fmt.Println("  leverage SYMBOL N                Set leverage for symbol")
	fmt.Println("  convert SYMBOL USDT LEV          Convert USDT to contract size")
	fmt.Println("  cancel-all SYMBOL                Cancel all orders")
	fmt.Println("  cancel-all-positions SYMBOL      Cancel all positions for a symbol (including open orders)")
	fmt.Println("  ticker SYMBOL                    Show ticker info")
	fmt.Println("  orderbook SYMBOL                 Show orderbook")
	fmt.Println("  orders SYMBOL                    Show active orders")
	fmt.Println("  fills SYMBOL                     Show fills")
	fmt.Println("  klines SYMBOL RESOLUTION         Show klines")
	fmt.Println()
}

func printJSON(data any) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("JSON error:", err)
		return
	}
	fmt.Println(string(b))
}

func printPositions(pos *connectors.GAccountPositions) {
	fmt.Printf("USDT Balance: %s\n", pos.Account.AccountBalanceRv)

	found := false

	for _, p := range pos.Positions {
		if p.SizeRq == "" || p.SizeRq == "0" {
			continue
		}

		found = true
		fmt.Println("------ OPEN POSITION ------")
		fmt.Printf("Symbol:     %s\n", p.Symbol)
		fmt.Printf("PosSide:    %s\n", p.PosSide)
		fmt.Printf("SizeRq:     %s\n", p.SizeRq)
		fmt.Printf("AvgPrice:   %s\n", p.AvgEntryPriceRp)
		fmt.Printf("Margin:     %s\n", p.PositionMarginRv)
		fmt.Printf("MarkPrice:  %s\n", p.MarkPriceRp)
		fmt.Println("---------------------------")
	}

	if !found {
		fmt.Println("No open USDT-M positions.")
	}
}

func printOrders(data json.RawMessage) {
	var payload struct {
		Rows    []map[string]interface{} `json:"rows"`
		HasNext bool                     `json:"hasNext"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		fmt.Println("Error parsing orders:", err)
		printJSON(data)
		return
	}

	if len(payload.Rows) == 0 {
		fmt.Println("No active orders.")
		return
	}

	for i, row := range payload.Rows {
		fmt.Printf("------ ORDER %d ------\n", i+1)
		printMapField(row, "symbol", "Symbol")
		printMapField(row, "side", "Side")
		printMapField(row, "posSide", "PosSide")
		printMapField(row, "ordType", "OrdType")
		printMapField(row, "priceRp", "PriceRp")
		printMapField(row, "orderQtyRq", "QtyRq")
		printMapField(row, "reduceOnly", "ReduceOnly")
		printMapField(row, "ordStatus", "Status")
		printMapField(row, "clOrdID", "ClientID")
		printMapField(row, "cumQtyRq", "FilledQty")
		printMapField(row, "leavesQtyRq", "LeavesQty")
		printMapField(row, "stopPxRp", "StopPx")
		fmt.Println("---------------------")
	}

	if payload.HasNext {
		fmt.Println("More orders available...")
	}
}

func printOrderbook(data json.RawMessage) {
	var payload struct {
		Depth      int    `json:"depth"`
		Dts        int64  `json:"dts"`
		Mts        int64  `json:"mts"`
		Timestamp  int64  `json:"timestamp"`
		Sequence   int64  `json:"sequence"`
		Symbol     string `json:"symbol"`
		Type       string `json:"type"`
		OrderbookP struct {
			Asks [][]string `json:"asks"`
			Bids [][]string `json:"bids"`
		} `json:"orderbook_p"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		fmt.Println("Error parsing orderbook:", err)
		printJSON(data)
		return
	}

	fmt.Println("------ ORDERBOOK ------")
	fmt.Printf("Symbol: %s\n", payload.Symbol)
	fmt.Printf("Timestamp: %d\n", payload.Timestamp)

	fmt.Println("Asks:")
	for _, lvl := range payload.OrderbookP.Asks {
		fmt.Printf("  Price: %s  Qty: %s\n", lvl[0], lvl[1])
	}

	fmt.Println("Bids:")
	for _, lvl := range payload.OrderbookP.Bids {
		fmt.Printf("  Price: %s  Qty: %s\n", lvl[0], lvl[1])
	}

	fmt.Println("-----------------------")
}

func printLevels(label string, levels [][]json.RawMessage) {
	fmt.Printf("%s (top 5):\n", label)
	if len(levels) == 0 {
		fmt.Println("  none")
		return
	}

	limit := 5
	if len(levels) < limit {
		limit = len(levels)
	}

	for i := 0; i < limit; i++ {
		var parts []string
		for _, raw := range levels[i] {
			parts = append(parts, strings.Trim(string(raw), "\""))
		}
		fmt.Printf("  %d) %s\n", i+1, strings.Join(parts, " | "))
	}
}

func printMapField(m map[string]interface{}, key, label string) {
	if v, ok := m[key]; ok {
		fmt.Printf("%-11s: %v\n", label, v)
	}
}

func main() {

	SetupLogger()
	apiKey := os.Getenv("KUCOIN_API_KEY")
	apiSecret := os.Getenv("KUCOIN_API_SECRET")
	apiPassphrase := os.Getenv("KUCOIN_API_PASSPHRASE")
	keyVersion := os.Getenv("KUCOIN_API_KEY_VERSION")

	//keyVersionInt, err := strconv.Atoi(keyVersion)
	//if err != nil {
	//	keyVersionInt = 3
	//}
	//
	//c := ccxt.Credentials{
	//	ApiKey:     apiKey,
	//	Secret:     apiSecret,
	//	Passphrase: apiPassphrase,
	//	KeyVersion: keyVersionInt,
	//}

	//type Credentials struct {
	//	ApiKey     string
	//	Secret     string
	//	Passphrase string
	//	KeyVersion int
	//}

	if apiKey == "" || apiSecret == "" {
		logger.Fatal("Missing API keys")
	}

	client := connectors.NewKucoinConnector(apiKey, apiSecret, apiPassphrase, keyVersion)

	reader := bufio.NewScanner(os.Stdin)
	// Increase the buffer so Ctrl+V pastes (especially long commands) are accepted without truncation.
	reader.Buffer(make([]byte, 0, 1024), 1024*1024)
	fmt.Println("Kucoin CLI Ready. Type 'help' for a list of commands. Type 'shutdown' to exit.")
	//price, err := strconv.ParseFloat("90", 64)
	for {
		fmt.Print("Kucoin> ")

		if !reader.Scan() {
			continue
		}

		line := strings.TrimSpace(reader.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, " ")
		cmd := parts[0]

		switch cmd {

		case "shutdown":
			fmt.Println("Exiting CLI...")
			return

		case "help":
			printUsage()

		case "positions":
			balances, err := client.GetAccountBalances()
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printJSON(balances)

		case "long":
			if len(parts) < 3 {
				printUsage()
				continue
			}
			symbol, qtyStr := parts[1], parts[2]
			size, err := strconv.ParseInt(qtyStr, 10, 64)
			if err != nil {
				fmt.Printf("Invalid size %q: %v\n", qtyStr, err)
				continue
			}

			fmt.Printf("Executing LONG %s size=%d\n", symbol, size)
			resp, err := client.ExecuteFuturesOrder(symbol, "buy", "market", size, nil, "1", false)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printJSON(resp)

		case "short":
			if len(parts) < 3 {
				printUsage()
				continue
			}
			symbol, qtyStr := parts[1], parts[2]
			size, err := strconv.ParseInt(qtyStr, 10, 64)
			if err != nil {
				fmt.Printf("Invalid size %q: %v\n", qtyStr, err)
				continue
			}

			fmt.Printf("Executing SHORT %s size=%d\n", symbol, size)
			resp, err := client.ExecuteFuturesOrder(symbol, "sell", "market", size, nil, "1", false)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printJSON(resp)

		case "close-long":
			if len(parts) < 3 {
				printUsage()
				continue
			}
			symbol, qtyStr := parts[1], parts[2]
			size, err := strconv.ParseInt(qtyStr, 10, 64)
			if err != nil {
				fmt.Printf("Invalid size %q: %v\n", qtyStr, err)
				continue
			}

			fmt.Printf("Closing LONG %s size=%d\n", symbol, size)
			resp, err := client.ExecuteFuturesOrder(symbol, "sell", "market", size, nil, "1", true)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printJSON(resp)

		case "close-short":
			if len(parts) < 3 {
				printUsage()
				continue
			}
			symbol, qtyStr := parts[1], parts[2]
			size, err := strconv.ParseInt(qtyStr, 10, 64)
			if err != nil {
				fmt.Printf("Invalid size %q: %v\n", qtyStr, err)
				continue
			}

			fmt.Printf("Closing SHORT %s size=%d\n", symbol, size)
			resp, err := client.ExecuteFuturesOrder(symbol, "buy", "market", size, nil, "1", true)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printJSON(resp)

		case "reverse":
			if len(parts) < 3 {
				printUsage()
				continue
			}
			symbol, qtyStr := parts[1], parts[2]
			size, err := strconv.ParseInt(qtyStr, 10, 64)
			if err != nil {
				fmt.Printf("Invalid size %q: %v\n", qtyStr, err)
				continue
			}

			fmt.Printf("Reversing position on %s with size=%d\n", symbol, size)
			if _, err := client.ExecuteFuturesOrder(symbol, "sell", "market", size, nil, "1", true); err != nil {
				fmt.Println("Error closing existing position:", err)
				continue
			}
			resp, err := client.ExecuteFuturesOrder(symbol, "sell", "market", size, nil, "1", false)
			if err != nil {
				fmt.Println("Error opening reverse position:", err)
				continue
			}
			printJSON(resp)

		case "long-usdt", "short-usdt":
			if len(parts) < 4 {
				printUsage()
				continue
			}
			symbol := parts[1]
			usdt, err := strconv.ParseFloat(parts[2], 64)
			if err != nil || usdt <= 0 {
				fmt.Printf("Invalid USDT amount %q: %v\n", parts[2], err)
				continue
			}
			lev, err := strconv.Atoi(parts[3])
			if err != nil || lev <= 0 {
				fmt.Printf("Invalid leverage %q: %v\n", parts[3], err)
				continue
			}

			size, used, err := client.ConvertUSDTToContracts(symbol, usdt, lev)
			if err != nil {
				fmt.Println("Conversion error:", err)
				continue
			}

			fmt.Printf("Using %.4f USDT at %dx leverage gives %d contracts.\n", used, lev, size)
			if size == 0 {
				fmt.Println("Computed size is zero, aborting order.")
				continue
			}

			side := "buy"
			if cmd == "short-usdt" {
				side = "sell"
			}

			resp, err := client.ExecuteFuturesOrderLeverage(symbol, side, "market", size, nil, lev, false)
			if err != nil {
				fmt.Println("Error placing leveraged order:", err)
				continue
			}

			printJSON(resp)

		case "leverage":
			if len(parts) < 3 {
				printUsage()
				continue
			}
			symbol := parts[1]
			lev, err := strconv.Atoi(parts[2])
			if err != nil || lev <= 0 {
				fmt.Printf("Invalid leverage %q: %v\n", parts[2], err)
				continue
			}

			if err := client.SetFuturesLeverage(symbol, lev); err != nil {
				fmt.Println("Error setting leverage:", err)
				continue
			}

			fmt.Printf("Leverage for %s set to %dx.\n", symbol, lev)

		case "convert":
			if len(parts) < 4 {
				printUsage()
				continue
			}
			symbol := parts[1]
			usdt, err := strconv.ParseFloat(parts[2], 64)
			if err != nil || usdt <= 0 {
				fmt.Printf("Invalid USDT amount %q: %v\n", parts[2], err)
				continue
			}
			lev, err := strconv.Atoi(parts[3])
			if err != nil || lev <= 0 {
				fmt.Printf("Invalid leverage %q: %v\n", parts[3], err)
				continue
			}

			size, used, err := client.ConvertUSDTToContracts(symbol, usdt, lev)
			if err != nil {
				fmt.Println("Error converting:", err)
				continue
			}

			fmt.Printf("%s => contracts: %d (using %.6f USDT at %dx)\n", symbol, size, used, lev)

		case "cancel-all":
			fmt.Println("Cancel-all is not supported by the KuCoin connector yet.")

		case "cancel-all-positions":
			fmt.Println("Cancel-all-positions is not supported by the KuCoin connector yet.")

		case "ticker":
			if len(parts) < 2 {
				printUsage()
				continue
			}
			symbol := parts[1]
			data, err := client.GetFuturesTicker(symbol)
			//resp, err := client.GetTicker(symbol)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printJSON(data)

			data1, err := client.GetFuturesContractInfo(symbol)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printJSON(data1)

			data3, err := client.GetFuturesContractInfoRaw(symbol)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printJSON(data3)

		case "orderbook":
			//if len(parts) < 2 {
			//	printUsage()
			//	continue
			//}
			//symbol := parts[1]
			//resp, err := client.GetOrderbook(symbol)
			//if err != nil {
			//	fmt.Println("Error:", err)
			//	continue
			//}
			//printOrderbook(resp.Data)

		case "orders":
			//if len(parts) < 2 {
			//	printUsage()
			//	continue
			//}
			//symbol := parts[1]
			//resp, err := client.GetActiveOrders(symbol)
			//if err != nil {
			//	fmt.Println("Error:", err)
			//	continue
			//}
			//printOrders(resp.Data)

		case "ordershistory":
			//if len(parts) < 2 {
			//	printUsage()
			//	continue
			//}
			//symbol := parts[1]
			//resp, err := client.GetOrderHistory(symbol)
			//if err != nil {
			//	fmt.Println("Error:", err)
			//	continue
			//}
			//printOrders(resp.Data)

		case "fills":
			//if len(parts) < 2 {
			//	printUsage()
			//	continue
			//}
			//symbol := parts[1]
			//resp, err := client.GetFills(symbol)
			//if err != nil {
			//	fmt.Println("Error:", err)
			//	continue
			//}
			//printJSON(resp.Data)

		case "klines":
			//if len(parts) < 3 {
			//	printUsage()
			//	continue
			//}
			//symbol := parts[1]
			//res, _ := strconv.Atoi(parts[2])
			//resp, err := client.GetKlines(symbol, res)
			//if err != nil {
			//	fmt.Println("Error:", err)
			//	continue
			//}
			//printJSON(resp.Data)

		case "disp":
			if len(parts) < 2 {
				printUsage()
				continue
			}
			symbol := parts[1]
			avail, err := client.GetFuturesAvailableForSymbol(symbol)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			fmt.Printf("USDT available for %s: %.12f\n", symbol, avail)

		case "avl":

			pos, err := client.GetAccountBalances()
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printJSON(pos)

			if err := client.TestConnection(); err != nil {
				fmt.Println("Error:", err)
				continue
			}

			avail, err := client.GetFuturesAvailableForSymbol("XBTUSDTM")
			if err != nil {
				// handle error
			}
			fmt.Println("Available USDT for futures:", avail)

		default:
			fmt.Println("Unknown command:", cmd)
			printUsage()
		}
	}
}
