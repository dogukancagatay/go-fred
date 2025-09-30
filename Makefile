# Go-Fred-REST Makefile

.PHONY: build test test-verbose test-coverage clean run help

# Build the application
build:
	go build -o go-fred-rest .

# Run tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run tests with coverage and show coverage percentage
test-coverage-summary:
	go test -cover ./...

# Clean build artifacts
clean:
	rm -f go-fred-rest
	rm -f coverage.out coverage.html

# Run the application
run: build
	./go-fred-rest

# Run the application with hot reload (requires air)
dev:
	air

# Install development dependencies
install-dev:
	go install github.com/cosmtrek/air@latest

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Install linting tools
install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run all checks (format, lint, test)
check: fmt lint test

# Help
help:
	@echo "Available targets:"
	@echo "  build              - Build the application"
	@echo "  test               - Run tests"
	@echo "  test-verbose       - Run tests with verbose output"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo "  test-coverage-summary - Run tests with coverage summary"
	@echo "  clean              - Clean build artifacts"
	@echo "  run                - Build and run the application"
	@echo "  dev                - Run with hot reload (requires air)"
	@echo "  install-dev        - Install development dependencies"
	@echo "  fmt                - Format code"
	@echo "  lint               - Lint code"
	@echo "  install-lint       - Install linting tools"
	@echo "  check              - Run all checks (format, lint, test)"
	@echo "  help               - Show this help message"
