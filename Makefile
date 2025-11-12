.PHONY: lint format test up up-build down logs build-server run-server build-client run-client

# Start docker compose services
up:
	@echo "Starting docker compose services..."
	@docker compose up

# Start docker compose services with build
up-build:
	@echo "Building and starting docker compose services..."
	@docker compose up --build

# Stop docker compose services
down:
	@echo "Stopping docker compose services..."
	@docker compose down

# View docker compose logs
logs:
	@docker compose logs -f

# Build the server binary
build-server:
	@echo "Building server..."
	@cd server && make build

# Run the server locally
run-server:
	@echo "Starting server on port 8080..."
	@cd server && PORT=8080 go run ./cmd/server

# Build the client
build-client:
	@echo "Building client..."
	@cd client && npm run build

# Run the client dev server
run-client:
	@echo "Starting client dev server on port 3000..."
	@cd client && npm run dev

# Lint all code
lint:
	@echo "Linting Go code..."
	@cd server && golangci-lint run ./...
	@echo "Linting TypeScript code..."
	@cd client && npm run lint

# Format all code
format:
	@echo "Formatting Go code..."
	@cd server && gofmt -s -w . && goimports -w .
	@echo "Formatting TypeScript code..."
	@cd client && npm run format

# Run all tests
test:
	@echo "Running Go tests..."
	@cd server && go test ./...
	@echo "Running TypeScript tests..."
	@cd client && npm test -- --run

