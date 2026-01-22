BINARY_NAME := cas
CMD_PATH := ./cmd/main
BUILD_DIR := ./bin

.PHONY: build run clean test test-v test-race test-cover fmt vet check serve help

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)

run:
	go run $(CMD_PATH)

clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html 

test:
	go test ./pkg/...

test-v:
	go test -v ./pkg/...

test-race:
	go test -v -race ./pkg/...

test-cover:
	go test -coverprofile=coverage.out ./pkg/...
	go tool cover -html=coverage.out -o coverage.html

fmt:
	gofmt -s -w .

vet:
	go vet ./...

check: fmt vet

serve: build
	$(BUILD_DIR)/$(BINARY_NAME) serve

help:                                                                    
	@echo "  build       Build the binary"                                                  
	@echo "  run         Run with go run"                                                   
	@echo "  clean       Remove artifacts"                                                  
	@echo "  test        Run all tests"                                                     
	@echo "  test-v      Verbose tests"                                                     
	@echo "  test-race   Tests with race detector"                                          
	@echo "  test-cover  Coverage report"                                                   
	@echo "  fmt         Format code"                                                       
	@echo "  vet         Static analysis"                                                   
	@echo "  check       Run fmt + vet"                                                     
	@echo "  serve       Build and start server"                                                                                       
	@echo "  help        Show this help"