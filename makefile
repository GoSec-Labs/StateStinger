.PHONY: build test clean install docker run help

# Project variables
BINARY_NAME=statestinger
DOCKER_IMAGE=statestinger
GO_FILES=$(shell find . -name "*.go" -type f)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build settings
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/statestinger

# Install the binary
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) ./cmd/statestinger

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
cover:
	@echo "Running tests with coverage..."
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -s -w $(GO_FILES)

# Lint code
lint:
	@echo "Linting code..."
	golint ./...
	go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Build Docker image
docker:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

# Run in Docker
docker-run:
	@echo "Running in Docker..."
	docker run -v $(PWD)/results:/data $(DOCKER_IMAGE) $(ARGS)

# Show help
help:
	@echo "StateStinger Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build      Build the binary"
	@echo "  install    Install the binary"
	@echo "  test       Run tests"
	@echo "  cover      Run tests with coverage"
	@echo "  fmt        Format code"
	@echo "  lint       Lint code"
	@echo "  clean      Clean build artifacts"
	@echo "  docker     Build Docker image"
	@echo "  docker-run Run in Docker (use ARGS=\"--target /path --output /data\")"
	@echo "  help       Show this help"