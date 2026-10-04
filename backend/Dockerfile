# --- Build Stage ---
FROM golang:alpine AS builder

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -ldflags="-w -s" -o /bin/api ./cmd/api

# --- Runtime Stage ---
FROM alpine:3.21

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /bin/api /app/api

EXPOSE 8080

CMD ["/app/api"]
