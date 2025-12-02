package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"vsC1Y2025V01/internal/connectors"
)

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

func main() {
	apiKey := os.Getenv("PHEMEX_API_KEY")
	apiSecret := os.Getenv("PHEMEX_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		log.Fatal("Missing API keys")
	}

	client := connectors.NewClient(apiKey, apiSecret)

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
			printJSON(pos)

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
			printJSON(resp.Data)

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
			printJSON(resp.Data)

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

		default:
			fmt.Println("Unknown command:", cmd)
			printUsage()
		}
	}
}
