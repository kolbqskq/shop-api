
include .env
export

service-run:
	go run ./cmd

migrate-up:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres ${DATABASE_URL} up

migrate-down:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres ${DATABASE_URL} down-to 0