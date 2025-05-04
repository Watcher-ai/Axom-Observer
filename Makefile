.PHONY: all build test lint fmt run

all: build

build:
	go build -o observer ./cmd/observer

test:
	go test ./...

lint:
	golangci-lint run

fmt:
	go fmt ./...

run:
	sudo ./observer

docker-build:
	docker build -t axom-observer .

docker-up:
	docker-compose up --build

clean:
	rm -f observer
