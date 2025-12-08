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
	apiKey := os.Getenv("PHEMEX_API_KEY")
	apiSecret := os.Getenv("PHEMEX_API_SECRET")
	baseURL := os.Getenv("PHEMEX_BASE_URL")

	if apiKey == "" || apiSecret == "" {
		logger.Fatal("Missing API keys")
	}

	client := connectors.NewClient(apiKey, apiSecret, baseURL)

	reader := bufio.NewScanner(os.Stdin)
	fmt.Println("Phemex CLI Ready. Type 'help' for a list of commands. Type 'shutdown' to exit.")

	for {
		fmt.Print("phemex> ")

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
			pos, err := client.GetPositionsUSDT()
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printPositions(pos)

		case "long":
			if len(parts) < 3 {
				printUsage()
				continue
			}
			symbol, qty := parts[1], parts[2]
			fmt.Printf("Executing LONG %s qty=%s\n", symbol, qty)
			resp, err := client.PlaceOrder(symbol, "Buy", "Long", qty, "Market", false)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printJSON(resp.Data)

		case "short":
			if len(parts) < 3 {
				printUsage()
				continue
			}
			symbol, qty := parts[1], parts[2]
			fmt.Printf("Executing SHORT %s qty=%s\n", symbol, qty)
			resp, err := client.PlaceOrder(symbol, "Sell", "Short", qty, "Market", false)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printJSON(resp.Data)

		case "close-long":
			if len(parts) < 3 {
				printUsage()
				continue
			}
			symbol, qty := parts[1], parts[2]
			fmt.Printf("Closing LONG %s qty=%s\n", symbol, qty)
			resp, err := client.PlaceOrder(symbol, "Sell", "Long", qty, "Market", true)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printJSON(resp.Data)

		case "close-short":
			if len(parts) < 3 {
				printUsage()
				continue
			}
			symbol, qty := parts[1], parts[2]
			fmt.Printf("Closing SHORT %s qty=%s\n", symbol, qty)
			resp, err := client.PlaceOrder(symbol, "Buy", "Short", qty, "Market", true)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printJSON(resp.Data)

		case "reverse":
			if len(parts) < 3 {
				printUsage()
				continue
			}
			symbol, qty := parts[1], parts[2]
			fmt.Printf("Reversing %s qty=%s\n", symbol, qty)

			client.PlaceOrder(symbol, "Sell", "Long", qty, "Market", true)
			resp, err := client.PlaceOrder(symbol, "Sell", "Short", qty, "Market", false)

			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printJSON(resp.Data)

		case "cancel-all":
			if len(parts) < 2 {
				printUsage()
				continue
			}
			symbol := parts[1]
			resp, err := client.CancelAll(symbol)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printJSON(resp.Data)

		case "cancel-all-positions":
			if len(parts) < 2 {
				printUsage()
				continue
			}
			symbol := parts[1]
			err := client.CloseAllPositions(symbol)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			pos, err := client.GetPositionsUSDT()
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printPositions(pos)

		case "ticker":
			if len(parts) < 2 {
				printUsage()
				continue
			}
			symbol := parts[1]
			resp, err := client.GetTicker(symbol)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printJSON(resp.Data)

		case "orderbook":
			if len(parts) < 2 {
				printUsage()
				continue
			}
			symbol := parts[1]
			resp, err := client.GetOrderbook(symbol)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printOrderbook(resp.Data)

		case "orders":
			if len(parts) < 2 {
				printUsage()
				continue
			}
			symbol := parts[1]
			resp, err := client.GetActiveOrders(symbol)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printOrders(resp.Data)

		case "ordershistory":
			if len(parts) < 2 {
				printUsage()
				continue
			}
			symbol := parts[1]
			resp, err := client.GetOrderHistory(symbol)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printOrders(resp.Data)

		case "fills":
			if len(parts) < 2 {
				printUsage()
				continue
			}
			symbol := parts[1]
			resp, err := client.GetFills(symbol)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printJSON(resp.Data)

		case "klines":
			if len(parts) < 3 {
				printUsage()
				continue
			}
			symbol := parts[1]
			res, _ := strconv.Atoi(parts[2])
			resp, err := client.GetKlines(symbol, res)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printJSON(resp.Data)

		case "disp":
			if len(parts) < 2 {
				printUsage()
				continue
			}
			symbol := parts[1]

			qtd, err := client.GetFuturesAvailableFromRiskUnit(symbol)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}

			fmt.Printf("USDT available %.12f\n", qtd)

		case "avl":
			if len(parts) < 2 {
				printUsage()
				continue
			}
			symbol := parts[1]

			baseSymbol, baseAvail, usdtAvail, price, err := client.GetAvailableBaseFromUSDT(symbol)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}

			fmt.Printf("Available %s\n", baseSymbol)
			fmt.Printf("USDT USDT -> base coin %.12f\n", baseAvail)
			fmt.Printf("USDT available %.12f\n", usdtAvail)
			fmt.Printf("USDT price %.12f\n", price)

		default:
			fmt.Println("Unknown command:", cmd)
			printUsage()
		}
	}
}
