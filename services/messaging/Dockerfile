FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY shared/ ./shared/
COPY services/messaging/ ./services/messaging/
RUN CGO_ENABLED=0 GOOS=linux go build -o messaging ./services/messaging

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/messaging .
EXPOSE 8085
CMD ["./messaging"]
