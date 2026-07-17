FROM golang:1.26 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build \
    -o user \
    -ldflags="-s -w" \
    ./services/user

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/user .
EXPOSE 8081
CMD ["./user"]
