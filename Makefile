.PHONY: build clean test run help

# Build variables
BINARY_NAME=goclaw
BUILD_DIR=bin
GO_FILES=$(shell find . -name '*.go' -type f)

# Default target
all: build

build: $(BUILD_DIR)/$(BINARY_NAME)

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES)
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/goclaw/
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f *.db
	@echo "Clean complete"

test:
	@echo "Running tests..."
	go test -v ./...

run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME) chat

install: build
	@echo "Installing $(BINARY_NAME) to ~/go/bin/..."
	@mkdir -p ~/go/bin
	cp $(BUILD_DIR)/$(BINARY_NAME) ~/go/bin/

help:
	@echo "Available targets:"
	@echo "  build    - Build the binary"
	@echo "  clean    - Remove build artifacts"
	@echo "  test     - Run tests"
	@echo "  run      - Build and run chat mode"
	@echo "  install  - Install to ~/go/bin/"
	@echo "  help     - Show this help message"
