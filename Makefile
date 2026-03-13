.PHONY: all lint vet build test clean

BINARY := go-kusto-cli
GOBIN  := $(shell go env GOPATH)/bin

all: lint vet test build

lint:
	$(GOBIN)/golangci-lint run ./...

vet:
	go vet ./...

build:
	go build -o $(BINARY) .

test:
	go test -race -count=1 ./...

clean:
	rm -f $(BINARY)
