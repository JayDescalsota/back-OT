.PHONY: run build test generate

run:
	go run .

build:
	go build -o bin/tenant-svc .

test:
	go test ./... -v

generate:
	go generate ./...

migrate-up:
	go run github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
		-path migrations \
		-database "$$TENANTDB_URL" \
		up

migrate-down:
	go run github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
		-path migrations \
		-database "$$TENANTDB_URL" \
		down 1
