# trading-journal
# NexAPI Integration Project

This project demonstrates how to integrate with multiple cryptocurrency exchanges using the [NexAPI](https://github.com/linstohu/nexapi) library. It provides a unified interface for interacting with various exchange APIs for operations like testing connections, fetching account balances, and placing orders.

## Features
- Connect to multiple exchanges supported by NexAPI.
- Fetch account balances and perform basic trading operations.
- Modular structure for scalability and maintainability.

## Project Structure
```
vsC1Y2025V01/
├── cmd/
│   └── main.go                 # Entry point of the application
├── internal/
│   ├── connectors/
│   │   ├── mexc_connector.go   # Mexc connector implementation
│   │   ├── kucoin_connector.go # KuCoin connector implementation
│   │   └── connector_interface.go # Unified connector interface
│   └── tests/
│       ├── mexc_connector_test.go # Unit tests for Mexc connector
│       ├── kucoin_connector_test.go # Unit tests for KuCoin connector
│       └── connector_interface_test.go # Tests for the connector interface
├── pkg/
│   └── utils/
│       └── logger.go           # Logging utilities (placeholder)
├── Makefile                    # Automation tasks (build, test, etc.)
├── .env                        # Environment variables (API keys, etc.)
├── .gitignore                  # Ignored files for Git
├── README.md                   # Project documentation (this file)
```

## Prerequisites
- Go 1.18+
- API keys for the exchanges you want to connect to (configured in `.env`).

## Installation
1. Clone this repository:
   ```bash
   git clone https://github.com/your-repo/vsC1Y2025V01.git
   cd vsC1Y2025V01
   ```
2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Usage
1. Update the `.env` file with your API keys.
2. Run the application:
   ```bash
   go run ./cmd/main.go
   ```
3. Run tests:
   ```bash
   go test ./...
   ```

## License
This project is licensed under the MIT License.

## References

## Run the migration

   ```bash
    export DATABASE_URL="postgres://user:password@localhost:5432/yourdb?sslmode=disable"
migrate -database "$DATABASE_URL" -path db/migrations up
# to check version
migrate -database "$DATABASE_URL" -path db/migrations version
# to rollback one step (if needed)
# migrate -database "$DATABASE_URL" -path db/migrations down 1

   ```

If you use Docker Compose, you can run the CLI inside a Postgres or app container; or call migrations from a Makefile/CI step.



