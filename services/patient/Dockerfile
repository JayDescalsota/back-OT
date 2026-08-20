FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY shared/ ./shared/
COPY services/patient/ ./services/patient/
RUN CGO_ENABLED=0 GOOS=linux go build -o patient ./services/patient

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/patient .
EXPOSE 8183
CMD ["./patient"]