FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY shared/ ./shared/
COPY services/gateway/ ./services/gateway/
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway ./services/gateway

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/gateway .
EXPOSE 4000
CMD ["./gateway"]
