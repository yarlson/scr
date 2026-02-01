.PHONY: all build test lint clean

all: lint test build

build:
	go build -o tui-capture ./cmd

test:
	go test -v -race -timeout=30s ./...

lint:
	golangci-lint run ./... --fix

clean:
	rm -f tui-capture
