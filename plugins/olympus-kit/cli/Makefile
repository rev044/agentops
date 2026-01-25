.PHONY: build test lint clean install

# Build variables
BINARY_NAME=ol
BUILD_DIR=bin
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
all: build

# Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/ol

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Install to GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/ol

# Run the binary
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Tidy go modules
tidy:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Generate (for future code generation)
generate:
	go generate ./...
