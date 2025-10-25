FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -o tsdb-server main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/tsdb-server .

RUN mkdir -p /app/data

EXPOSE 8080

CMD ["./tsdb-server", "-data-dir=/app/data", "-host=0.0.0.0", "-port=8080"]