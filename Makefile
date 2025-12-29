.PHONY: build run test clean docker-up docker-down

build:
	go build -o task-pool ./cmd/main.go

run:
	go run ./cmd/main.go http

test:
	go test ./...

clean:
	rm -f task-pool
	go clean

docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down
