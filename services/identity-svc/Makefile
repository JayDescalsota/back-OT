.PHONY: run generate test test-race cover cover-html build

run:
	go run main.go

generate:
	gqlgen generate

test:
	go test ./... -count=1

test-race:
	go test ./... -race -count=1

cover:
	go test ./... -coverprofile=coverage.out -covermode=atomic -count=1
	go tool cover -func=coverage.out

cover-html:
	go test ./... -coverprofile=coverage.out -covermode=atomic -count=1
	go tool cover -html=coverage.out -o coverage.html

build:
	go build -o bin/server main.go
