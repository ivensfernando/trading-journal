include ./scripts/env.sh

APP_NAME=VSC1Y2025

run:
	@echo "Running the application..."
	go run ./cmd/main.go

build:
	@echo "Building the application..."
	go build -o $(APP_NAME) ./cmd/main.go

clean:
	@echo "Cleaning up..."
	rm -f $(APP_NAME)

test:
	@echo "Running tests..."
	go test ./... -v

env:
	@echo "Exporting environment variables..."
	export $(cat .env | xargs)


build_server:
	go build -o $(BINARY_NAME) ./...

run_server:
	go run .




