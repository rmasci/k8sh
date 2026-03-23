# Makefile for k8sh

.PHONY: test test-verbose test-race test-cover clean build build-all lint fmt vet test-posix demo-posix demo-help

# Default target
all: test build

# Run all tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with race detection
test-race:
	go test -race ./...

# Run tests with coverage
test-cover:
	go test -cover ./...

# Run tests with coverage report
test-cover-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run specific package tests
test-k8s:
	go test ./pkg/k8s/...

test-ops:
	go test ./pkg/ops/...

test-shell:
	go test ./pkg/shell/...

test-editor:
	go test ./pkg/editor/...

test-posix:
	go test ./pkg/posix/...

test-integration:
	go test ./tests/...

# Run POSIX compliance tests
test-posix-compliance:
	go test ./pkg/posix/ -v -tags=posix

# Run POSIX demo
demo-posix:
	go run examples/posix/demo.go

# Show help demo
demo-help:
	@echo "🐚 k8sh - Improved Help Demo"
	@echo "=================================="
	@echo ""
	@echo "First-time user experience - much more friendly!"
	@echo ""
	@echo "Here's what users now see when they type 'help':"
	@echo ""
	@echo "🚀 QUICK START:"
	@echo "  1. List available pods:     k8sh> pods"
	@echo "  2. Select a pod:           k8sh> use my-pod"
	@echo "  3. Start working:           k8sh> ls -la"
	@echo "  4. Get help anytime:        k8sh> help"
	@echo ""
	@echo "📁 FILE OPERATIONS:"
	@echo "  mkdir <path>           Create directory"
	@echo "    Example: mkdir /app/data"
	@echo "  rm [-r] <path>         Remove file/directory"
	@echo "    Example: rm -r /tmp/old-data"
	@echo "  cp <src> <dst>         Copy file within pod"
	@echo "    Example: cp config.yaml config.yaml.bak"
	@echo "  download <src> <dst>    🆕 Download file from pod to local"
	@echo "    Example: download /app/logs/app.log ./logs-backup.log"
	@echo ""
	@echo "💡 PRO TIPS:"
	@echo "  • Use 'download' to copy files from pod to your local machine"
	@echo "  • All file paths work with both absolute (/path) and relative (path) formats"
	@echo "  • Tab completion works for pod and container names"
	@echo ""
	@echo "🐚 POSIX MODE:"
	@echo "  For full POSIX compliance: k8sh posix"
	@echo "  • Command pipelines: cmd1 | cmd2 | cmd3"
	@echo "  • I/O redirection: >, >>, <, 2>"
	@echo "  • Variable expansion: \$$VAR, \$$\{VAR\}"
	@echo ""
	@echo "Happy container hacking! 🎉"

# Run benchmarks
bench:
	go test -bench=. ./...

# Build the application (current platform)
build:
	@mkdir -p releases
	go build -o releases/k8sh ./cmd/k8sh
	@echo "✅ Built: releases/k8sh"

# Build for multiple platforms
build-all:
	@echo "🏗 Building k8sh for all platforms..."
	@echo "=================================="
	@mkdir -p releases
	@echo "Building for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 go build -o releases/k8sh-darwin-amd64 ./cmd/k8sh
	@echo "Building for macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 go build -o releases/k8sh-darwin-arm64 ./cmd/k8sh
	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build -o releases/k8sh-linux-amd64 ./cmd/k8sh
	@echo "Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 go build -o releases/k8sh-linux-arm64 ./cmd/k8sh
	@echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build -o releases/k8sh-windows-amd64.exe ./cmd/k8sh
	@echo ""
	@echo "✅ Build complete! Files in releases/:"
	@ls -la releases/
	@echo ""
	@echo "📦 Distribution ready!"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf releases/
	@rm -f coverage.out coverage.html
	@echo "✅ Clean complete!"

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Run all quality checks
quality: fmt vet lint

# Install dependencies
deps:
	go mod download
	go mod tidy

# Update dependencies
update-deps:
	go get -u ./...
	go mod tidy

# Generate test coverage report for CI
ci-test:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Check for security vulnerabilities (requires gosec)
security:
	gosec ./...

# Run mutation testing (requires gremlins)
mutation:
	gremlins ./...

# Profile the application
profile:
	go build -o bin/k8sh-profile ./cmd/k8sh
	./bin/k8sh-profile -cpuprofile=cpu.prof -memprofile=mem.prof

# Documentation
docs:
	godoc -http=:6060

# Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install github.com/go-critic/go-critic/cmd/gocritic@latest
	go install github.com/google/go-gremlins/cmd/gremlins@latest

# Development setup
setup: install-tools deps

# CI pipeline
ci: quality test-race security

# Help target
help:
	@echo "Available targets:"
	@echo "  test          - Run all tests"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-race     - Run tests with race detection"
	@echo "  test-cover    - Run tests with coverage"
	@echo "  test-cover-html - Generate HTML coverage report"
	@echo "  test-k8s      - Run k8s package tests"
	@echo "  test-ops      - Run ops package tests"
	@echo "  test-shell    - Run shell package tests"
	@echo "  test-editor   - Run editor package tests"
	@echo "  test-posix    - Run POSIX package tests"
	@echo "  test-posix-compliance - Run POSIX compliance tests"
	@echo "  test-integration - Run integration tests"
	@echo "  bench         - Run benchmarks"
	@echo "  build         - Build for current platform (./bin/)"
	@echo "  build-all     - Build for all platforms (./releases/)"
	@echo "  clean         - Clean build artifacts"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  lint          - Run linter"
	@echo "  quality       - Run all quality checks"
	@echo "  deps          - Install dependencies"
	@echo "  update-deps   - Update dependencies"
	@echo "  ci-test       - Run tests for CI"
	@echo "  security      - Run security scan"
	@echo "  mutation      - Run mutation testing"
	@echo "  profile       - Profile the application"
	@echo "  docs          - Generate documentation"
	@echo "  install-tools - Install development tools"
	@echo "  setup         - Development setup"
	@echo "  ci            - CI pipeline"
	@echo "  help          - Show this help"
