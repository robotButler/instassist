.PHONY: build install uninstall clean test help

BINARY_NAME=inst
INSTALL_PREFIX=/opt/instassist
INSTALL_BIN=$(INSTALL_PREFIX)/$(BINARY_NAME)
SYMLINK_PATH=/usr/local/bin/$(BINARY_NAME)
SCHEMA_PATH=/usr/local/share/insta-assist
SUDO?=sudo
VERSION=1.0.0
GO_INSTALL_DIR?=$(shell go env GOBIN)
ifeq ($(strip $(GO_INSTALL_DIR)),)
  GO_INSTALL_DIR=$(shell go env GOPATH)/bin
endif

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	go build -ldflags "-X instassist.version=$(VERSION)" -o $(BINARY_NAME) ./cmd/inst
	@echo "Build complete: ./$(BINARY_NAME)"

install: build ## Build and install to system (/opt/instassist + symlink in /usr/local/bin)
	@if [ "$$EUID" -eq 0 ]; then echo "Please run 'make install' without sudo; it will prompt for your password when needed."; exit 1; fi
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PREFIX)..."
	$(SUDO) mkdir -p $(INSTALL_PREFIX)
	$(SUDO) cp $(BINARY_NAME) $(INSTALL_BIN)
	$(SUDO) chmod +x $(INSTALL_BIN)
	@echo "Linking $(SYMLINK_PATH) -> $(INSTALL_BIN)..."
	$(SUDO) ln -sf $(INSTALL_BIN) $(SYMLINK_PATH)
	@echo "Copying schema alongside binary and to $(SCHEMA_PATH)..."
	$(SUDO) cp options.schema.json $(INSTALL_PREFIX)/
	$(SUDO) mkdir -p $(SCHEMA_PATH)
	$(SUDO) cp options.schema.json $(SCHEMA_PATH)/
	@echo "Installation complete!"
	@echo ""
	@echo "To use the schema, the binary will look for it in:"
	@echo "  1. Same directory as the binary ($(INSTALL_PREFIX))"
	@echo "  2. Current working directory"
	@echo "  3. $(SCHEMA_PATH)/options.schema.json"

uninstall: ## Remove installed binary and schema
	@if [ "$$EUID" -eq 0 ]; then echo "Please run 'make uninstall' without sudo; it will prompt for your password when needed."; exit 1; fi
	@echo "Removing $(BINARY_NAME)..."
	$(SUDO) rm -f $(SYMLINK_PATH)
	$(SUDO) rm -f $(INSTALL_BIN)
	$(SUDO) rm -f $(INSTALL_PREFIX)/options.schema.json
	$(SUDO) rmdir $(INSTALL_PREFIX) 2>/dev/null || true
	@echo "Removing schema..."
	$(SUDO) rm -rf $(SCHEMA_PATH)
	@echo "Uninstall complete!"

user-install: build ## Install to ~/.local/bin (no sudo)
	@echo "Installing $(BINARY_NAME) to $$HOME/.local/bin..."
	mkdir -p $$HOME/.local/bin
	cp $(BINARY_NAME) $$HOME/.local/bin/
	@echo "Installation complete! Ensure $$HOME/.local/bin is on your PATH."

clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	@echo "Clean complete!"

test: build ## Build and run a quick test
	@echo "Testing version flag..."
	./$(BINARY_NAME) -version
	@echo ""
	@echo "Testing help..."
	./$(BINARY_NAME) -h

run: build ## Build and run in interactive mode
	./$(BINARY_NAME)

go-install: ## Install with go install (places binary in GOBIN or GOPATH/bin as inst)
	@echo "Installing to $(GO_INSTALL_DIR)"
	@mkdir -p "$(GO_INSTALL_DIR)"
	GOBIN=$(GO_INSTALL_DIR) go install ./cmd/inst
	@echo "Binary installed to $(GO_INSTALL_DIR)/inst"
