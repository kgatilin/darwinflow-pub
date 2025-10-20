.PHONY: build install test help

INSTALL_PATH := $(shell go env GOPATH)/bin

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build binary to ./dw
	go build -o dw ./cmd/dw

install: ## Install binary to GOPATH/bin
	go install ./cmd/dw
	@echo "Installed to: $(INSTALL_PATH)"

test: ## Run all tests
	go test ./...
