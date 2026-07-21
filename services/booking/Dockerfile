FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY shared/ ./shared/
COPY services/booking/ ./services/booking/
RUN CGO_ENABLED=0 GOOS=linux go build -o booking ./services/booking

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/booking .
EXPOSE 8084
CMD ["./booking"]