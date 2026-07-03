FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

RUN go build -o /app/bin/main ./cmd/app

# Финальный образ
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/main ./main
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/configs ./configs

RUN apk add --no-cache tzdata

EXPOSE 8080 9090

CMD ["./main"]