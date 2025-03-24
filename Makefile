.PHONY: build install test clean doc generate

# Go and binary settings
BINARY_NAME := terraform-provider-cloudcanary
VERSION := 0.1.0
OS_ARCH := $(shell go env GOOS)_$(shell go env GOARCH)

# The directory where Terraform looks for locally installed providers
LOCAL_PLUGIN_DIR := ~/.terraform.d/plugins/registry.terraform.io/yourorg/cloudcanary/$(VERSION)/$(OS_ARCH)

# Default target
all: build

# Build the provider binary
build:
	go build -o $(BINARY_NAME)

# Install the provider for local use
install: build
	mkdir -p $(LOCAL_PLUGIN_DIR)
	cp $(BINARY_NAME) $(LOCAL_PLUGIN_DIR)/

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -cover ./...

# Lint the code
lint:
	golangci-lint run ./...

# Generate documentation
generate:
	go generate ./...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

# Build and test
check: build test

# Development environment setup
dev-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest

# Documentation
doc: generate
	$(info Documentation generated)

# Run local terraform plan (useful for testing)
local-test: install
	cd examples && terraform init && terraform plan

# Build for all platforms (linux, darwin, windows)
build-all:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)_linux_amd64
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)_darwin_amd64
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)_windows_amd64.exe
