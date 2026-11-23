.PHONY: all build test lint clean run docker-build docker-up bench

BINARY_NAME=pulse-gateway

all: test build

build:
	go build -o bin/$(BINARY_NAME) -v ./cmd/gateway

test:
	go test -v -race ./...

bench:
	go test -bench=. -benchmem ./...

lint:
	go vet ./...

clean:
	rm -rf bin/ coverage.out

run: build
	./bin/$(BINARY_NAME) -config config/config.json

docker-build:
	docker build -t pulse-api-gateway:latest .

docker-up:
	docker-compose up -d --build
