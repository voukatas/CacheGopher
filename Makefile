
GO=go

BIN_DIR=bin

all: build_server build_cli

build_server:
	$(GO) build -o $(BIN_DIR)/server ./cmd/cachegopher/main.go
	cp cmd/cachegopher/cacheGopherConfig.json $(BIN_DIR)/cacheGopherConfig.json

build_cli:
	$(GO) build -o $(BIN_DIR)/gopher-cli ./cmd/cachegopher-cli/main.go

clean:
	rm -f $(BIN_DIR)/server $(BIN_DIR)/gopher-cli $(BIN_DIR)/cacheGopherConfig.json

test:
	go test -v -race -cover ./...

.PHONY: build_server build_cli test
