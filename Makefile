BINARY := hrzoner
DEVICES_BINARY := netdevices
MAIN := ./cmd/hrzoner
DEVICES_MAIN := ./cmd/netdevices
BUILD_DIR := bin
GOFLAGS ?= -buildvcs=false
GOCACHE ?= /tmp/hrzoner-go-cache

.PHONY: help run run-devices build build-devices test fmt vet tidy clean

help:
	@printf 'Available targets:\n'
	@printf '  run    Run the application\n'
	@printf '  run-devices    List devices on the local network\n'
	@printf '  build  Build the binary into $(BUILD_DIR)/$(BINARY)\n'
	@printf '  build-devices  Build the device scanner into $(BUILD_DIR)/$(DEVICES_BINARY)\n'
	@printf '  test   Run all tests\n'
	@printf '  fmt    Format Go files\n'
	@printf '  vet    Run go vet\n'
	@printf '  tidy   Tidy module dependencies\n'
	@printf '  clean  Remove build artifacts\n'

run:
	GOCACHE=$(GOCACHE) go run $(GOFLAGS) $(MAIN)

run-devices:
	GOCACHE=$(GOCACHE) go run $(GOFLAGS) $(DEVICES_MAIN)

build:
	mkdir -p $(BUILD_DIR)
	GOCACHE=$(GOCACHE) go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY) $(MAIN)

build-devices:
	mkdir -p $(BUILD_DIR)
	GOCACHE=$(GOCACHE) go build $(GOFLAGS) -o $(BUILD_DIR)/$(DEVICES_BINARY) $(DEVICES_MAIN)

test:
	GOCACHE=$(GOCACHE) go test $(GOFLAGS) ./...

fmt:
	GOCACHE=$(GOCACHE) go fmt ./...

vet:
	GOCACHE=$(GOCACHE) go vet $(GOFLAGS) ./...

tidy:
	GOCACHE=$(GOCACHE) go mod tidy

clean:
	rm -rf $(BUILD_DIR)
