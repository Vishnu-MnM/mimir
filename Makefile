.PHONY: all check test cover lint fmt tidy build

all: check

#check: lint + test (run this before every commit)
check: lint test

test:
	go test -race ./...

cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run

fmt:
	golangci-lint fmt

tidy:
	go mod tidy

# build: just checks if everything compiles; becomes -o bin/mimir ./cmd/mimir once /cmd exists.
build:
	go build ./...
