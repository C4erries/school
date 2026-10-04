# --- Build Stage ---
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Установка ca-certificates для HTTPS запросов
RUN apk --no-cache add ca-certificates git

# Кэширование зависимостей
COPY go.mod go.sum ./
RUN GODEBUG=netdns=go+4 go mod download

# Сборка приложения
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/api ./cmd/api

# --- Runtime Stage ---
FROM alpine:3.21

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /bin/api /app/api

EXPOSE 8080

CMD ["/app/api"]
