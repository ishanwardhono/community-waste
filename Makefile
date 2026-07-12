.PHONY: run test up down mock simulate migrate-down

run:
	go run ./cmd

test: mock
	go test ./...

up:
	docker compose up --build -d

down:
	docker compose down

mock:
	go generate ./...

simulate:
	./scripts/simulate.sh

migrate-down:
	docker compose run --rm migrate \
		-path /migrations -database "postgres://postgres:postgres@postgres:5432/community_waste?sslmode=disable" down 1
