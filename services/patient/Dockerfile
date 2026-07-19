FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o patient ./services/patient

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/patient .
EXPOSE 8083
CMD ["./patient"]