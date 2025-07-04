# Go parameters
GO_CMD=go
BINARY_NAME=gitversion-go

.PHONY: all build test clean run help

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@$(GO_CMD) build -o $(BINARY_NAME) ./cmd/gitversion-go

build-linux:
	@echo "Building statically linked binary for Linux..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO_CMD) build -o $(BINARY_NAME) ./cmd/gitversion-go

test:
	@echo "Running tests..."
	@$(GO_CMD) test ./...
	@/bin/zsh -c 'source ~/.zshrc && $(GO_CMD) test ./...'

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)

run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BINARY_NAME)

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build         - Build the application for the current OS"
	@echo "  build-linux   - Build a statically linked binary for Linux"
	@echo "  test     - Run tests"
	@echo "  clean    - Clean up build artifacts"
	@echo "  run      - Build and run the application"
	@echo "  help     - Show this help message"

.DEFAULT_GOAL := help
