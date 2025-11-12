.PHONY: build test test-verbose clean tidy

# Build the server binary
build:
	@echo "Building server..."
	@go build -o bin/server ./cmd/server

# Run all tests
test:
	@echo "Running tests..."
	@go test ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	@go test -v ./...

# Run tests with Ginkgo
test-ginkgo:
	@echo "Running tests with Ginkgo..."
	@ginkgo -r -v

# Run tests with specific labels (for contract tests)
test-labels:
	@echo "Running labeled tests..."
	@go test -v -tags=contract ./...

# Lint Go code
lint:
	@echo "Linting Go code..."
	@golangci-lint run ./...

# Format Go code
format:
	@echo "Formatting Go code..."
	@gofmt -s -w .
	@goimports -w .

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/

# Run go mod tidy
tidy:
	@echo "Running go mod tidy..."
	@go mod tidy

