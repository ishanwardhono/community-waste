.PHONY: run test up down mock

run:
	go run ./cmd

test:
	go test ./...

up:
	docker compose up --build -d

down:
	docker compose down

mock:
	go generate ./...
