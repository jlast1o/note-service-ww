.PHONY: run test swagger docker-build docker-up docker-down lint

run:
	go run ./cmd/server/

test:
	go test -v ./internal/handler/

swagger:
	swag init -g cmd/server/main.go

docker-build:
	docker build -t notes-service .

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down