.PHONY: run test lint migrate-up migrate-down swagger

run:
	go run ./cmd/api

test:
	go test ./...

lint:
	golangci-lint run ./...


swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go -o docs --parseDependency --parseInternal
