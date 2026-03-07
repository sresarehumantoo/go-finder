.PHONY: build test clean run lint fmt vet build-all

# Output directory
BIN_DIR := bin
APP     := go-finder

# Build all examples for the current platform
build:
	@mkdir -p $(BIN_DIR)
	@for dir in examples/*/; do \
		name=$$(basename $$dir); \
		go build -o $(BIN_DIR)/$$name ./$$dir; \
	done

# Cross-compile for all supported platforms
build-all: build-linux build-darwin build-windows build-freebsd
	@echo "All builds complete. Binaries in $(BIN_DIR)/"

# Linux
build-linux: build-linux-amd64 build-linux-arm64
build-linux-amd64:
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(APP)-linux-amd64 ./examples/basic
build-linux-arm64:
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=arm64 go build -o $(BIN_DIR)/$(APP)-linux-arm64 ./examples/basic

# macOS
build-darwin: build-darwin-amd64 build-darwin-arm64
build-darwin-amd64:
	@mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/$(APP)-darwin-amd64 ./examples/basic
build-darwin-arm64:
	@mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(BIN_DIR)/$(APP)-darwin-arm64 ./examples/basic

# Windows
build-windows: build-windows-amd64 build-windows-arm64
build-windows-amd64:
	@mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(APP)-windows-amd64.exe ./examples/basic
build-windows-arm64:
	@mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=arm64 go build -o $(BIN_DIR)/$(APP)-windows-arm64.exe ./examples/basic

# FreeBSD
build-freebsd: build-freebsd-amd64 build-freebsd-arm64
build-freebsd-amd64:
	@mkdir -p $(BIN_DIR)
	GOOS=freebsd GOARCH=amd64 go build -o $(BIN_DIR)/$(APP)-freebsd-amd64 ./examples/basic
build-freebsd-arm64:
	@mkdir -p $(BIN_DIR)
	GOOS=freebsd GOARCH=arm64 go build -o $(BIN_DIR)/$(APP)-freebsd-arm64 ./examples/basic

# Run all tests
test:
	go test -v -count=1 ./...

# Run tests with coverage
cover:
	@mkdir -p $(BIN_DIR)
	go test -coverprofile=$(BIN_DIR)/coverage.out -count=1 ./...
	go tool cover -html=$(BIN_DIR)/coverage.out -o $(BIN_DIR)/coverage.html
	@echo "Coverage report: $(BIN_DIR)/coverage.html"

# Run the basic example
run: build
	./$(BIN_DIR)/$(APP)

# Format code
fmt:
	gofmt -w .

# Vet code
vet:
	go vet ./...

# Lint (requires golangci-lint)
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf $(BIN_DIR)

# Build, vet, and test
all: fmt vet build test
