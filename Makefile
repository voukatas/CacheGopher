
GO=go

BIN_DIR=bin

all: build_server build_cli

build_server:
	$(GO) build -o $(BIN_DIR)/server ./cmd/cachegopher/main.go

build_cli:
	$(GO) build -o $(BIN_DIR)/gopher-cli ./cmd/cachegopher-cli/main.go

clean:
	rm -f $(BIN_DIR)/server $(BIN_DIR)/gopher-cli
build: 
	mkdir -p ./bin/
	go build -o ./bin/wky


.PHONY: build_server build_cli
