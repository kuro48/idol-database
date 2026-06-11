# ビルドステージ
FROM golang:1.26.4-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY cmd ./cmd
COPY internal ./internal
COPY static ./static

RUN swag init -g cmd/api/main.go -o docs

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o main ./cmd/api

# 実行ステージ
FROM alpine:3.24

RUN apk --no-cache add ca-certificates \
    && addgroup -S appgroup \
    && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/static ./static

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8081

CMD ["./main"]
