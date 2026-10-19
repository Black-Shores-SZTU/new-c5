.PHONY: run build fmt tidy test test-race cover golden up down logs

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

fmt:
	go fmt ./...

tidy:
	go mod tidy

test:
	go test ./...

test-race:
	CGO_ENABLED=1 go test -race ./...

cover:
	go test ./... -coverprofile=coverage.out

golden:
	go test ./internal/route -run TestGolden -update

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f api
