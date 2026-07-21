FROM golang:1.26 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY shared/ ./shared/
COPY services/tenant/ ./services/tenant/
RUN CGO_ENABLED=0 GOOS=linux go build -o tenant ./services/tenant

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/tenant .
EXPOSE 8082
CMD ["./tenant"]
