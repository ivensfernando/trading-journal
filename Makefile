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

swagger:
	@echo "Serving Swagger UI on http://localhost:8080 using swagger.yaml"
	docker run --rm -p 8080:8080 -e SWAGGER_JSON=swagger.yaml -v "$(PWD)/docs/swagger.yaml:/swagger.yaml" swaggerapi/swagger-ui


build_server:
	go build -o $(BINARY_NAME) ./...

run_server:
	go run .




