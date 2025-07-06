.PHONY: build run clean

BIN_DIR = bin
BINARY_NAME = server

build:
    @mkdir -p $(BIN_DIR)
    @go build -o $(BIN_DIR)/$(BINARY_NAME) cmd/server/main.go
    @echo "Build complete"

run: build
    @./$(BIN_DIR)/$(BINARY_NAME)

clean:
    @rm -rf $(BIN_DIR)
    @echo "Clean complete"