# Makefile for k8sh

.PHONY: test test-verbose test-race test-cover clean build lint fmt vet test-posix demo-posix

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

# Run benchmarks
bench:
	go test -bench=. ./...

# Build the application
build:
	go build -o bin/k8sh ./cmd/k8sh

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o bin/k8sh-linux-amd64 ./cmd/k8sh
	GOOS=darwin GOARCH=amd64 go build -o bin/k8sh-darwin-amd64 ./cmd/k8sh
	GOOS=windows GOARCH=amd64 go build -o bin/k8sh-windows-amd64.exe ./cmd/k8sh

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

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
	@echo "  build         - Build the application"
	@echo "  build-all     - Build for multiple platforms"
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
