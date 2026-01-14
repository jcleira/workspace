# Workspace Makefile

# Variables
BINARY_NAME=workspace
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"
GOFILES=$(shell find . -name "*.go" -type f)

# Colors
RED=\033[0;31m
GREEN=\033[0;32m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: all build clean test install uninstall fmt vet lint help completions

## help: Display this help message
help:
	@echo "Workspace Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  ${GREEN}%-15s${NC} %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

## all: Build the binary
all: build

## build: Build the binary for current platform
build:
	@echo "Building ${BINARY_NAME}..."
	@go build ${LDFLAGS} -o ${BINARY_NAME} .
	@echo "${GREEN}✓${NC} Build complete: ${BINARY_NAME}"

## build-all: Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p dist
	
	@echo "Building for Linux AMD64..."
	@GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}_linux_amd64 .
	
	@echo "Building for Linux ARM64..."
	@GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}_linux_arm64 .
	
	@echo "Building for macOS AMD64..."
	@GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}_darwin_amd64 .
	
	@echo "Building for macOS ARM64..."
	@GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}_darwin_arm64 .
	
	@echo "${GREEN}✓${NC} All builds complete"

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@go clean
	@rm -f ${BINARY_NAME}
	@rm -rf dist/
	@rm -rf completions/
	@echo "${GREEN}✓${NC} Clean complete"

## completions: Generate shell completion files
completions:
	@echo "Generating shell completions..."
	@go run scripts/generate-completions.go
	@echo "${GREEN}✓${NC} Completions generated in completions/"

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "${GREEN}✓${NC} Tests complete"

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "${GREEN}✓${NC} Coverage report generated: coverage.html"

## install: Install the binary to /usr/local/bin
install: build
	@echo "Installing ${BINARY_NAME} to /usr/local/bin..."
	@sudo mv ${BINARY_NAME} /usr/local/bin/
	@echo "${GREEN}✓${NC} Installation complete"
	@echo ""
	@echo "Run 'workspace config init' to set up shell integration"

## install-local: Install the binary to ~/go/bin
install-local: build
	@echo "Installing ${BINARY_NAME} to ~/go/bin..."
	@go install
	@echo "${GREEN}✓${NC} Installation complete"
	@echo ""
	@echo "Make sure ~/go/bin is in your PATH"

## uninstall: Remove the binary from /usr/local/bin
uninstall:
	@echo "Uninstalling ${BINARY_NAME}..."
	@sudo rm -f /usr/local/bin/${BINARY_NAME}
	@rm -f ~/.config/workspace/config.json
	@echo "${GREEN}✓${NC} Uninstall complete"

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "${GREEN}✓${NC} Formatting complete"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "${GREEN}✓${NC} Vet complete"

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	@GOBIN=$$(go env GOPATH | cut -d':' -f1)/bin; \
	if [ -x "$$GOBIN/golangci-lint" ]; then \
		$$GOBIN/golangci-lint run; \
		echo "${GREEN}✓${NC} Lint complete"; \
	elif command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
		echo "${GREEN}✓${NC} Lint complete"; \
	else \
		echo "${RED}✗${NC} golangci-lint not installed."; \
		echo "Install with: make lint-install"; \
	fi

## lint-fix: Run golangci-lint with auto-fix
lint-fix:
	@echo "Running linter with auto-fix..."
	@GOBIN=$$(go env GOPATH | cut -d':' -f1)/bin; \
	if [ -x "$$GOBIN/golangci-lint" ]; then \
		$$GOBIN/golangci-lint run --fix; \
		echo "${GREEN}✓${NC} Lint fix complete"; \
	elif command -v golangci-lint &> /dev/null; then \
		golangci-lint run --fix; \
		echo "${GREEN}✓${NC} Lint fix complete"; \
	else \
		echo "${RED}✗${NC} golangci-lint not installed."; \
		echo "Install with: make lint-install"; \
	fi

## lint-install: Install golangci-lint
lint-install:
	@echo "Installing golangci-lint v2.1.6..."
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
	@echo "${GREEN}✓${NC} golangci-lint v2.1.6 installed"

## check: Run fmt, vet, and lint
check: fmt vet lint
	@echo "${GREEN}✓${NC} All checks passed"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "${GREEN}✓${NC} Dependencies updated"

## update: Update dependencies
update:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "${GREEN}✓${NC} Dependencies updated"

## run: Run the application
run:
	@go run ${LDFLAGS} .

## release: Create a new release (requires VERSION parameter)
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "${RED}✗${NC} VERSION is required. Usage: make release VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "Creating release ${VERSION}..."
	@git tag -a ${VERSION} -m "Release ${VERSION}"
	@git push origin ${VERSION}
	@echo "${GREEN}✓${NC} Release ${VERSION} created"

## dev: Install for development with hot reload
dev:
	@echo "Installing for development..."
	@go install github.com/cosmtrek/air@latest
	@air init
	@air

.DEFAULT_GOAL := help