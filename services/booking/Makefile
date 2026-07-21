.PHONY: run build test generate migrate-up migrate-down

run:
	go run .

build:
	go build -o bin/booking-svc .

test:
	go test ./... -v

generate:
	go generate ./...

migrate-up:
	go run github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
		-path migrations \
		-database "$$BOOKINGDB_URL" \
		up

migrate-down:
	go run github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
		-path migrations \
		-database "$$BOOKINGDB_URL" \
		down 1
