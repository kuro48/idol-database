package logger

import (
	"context"
	"log/slog"
	"os"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// Setup はグローバルslogロガーをJSON形式で初期化する
func Setup(level slog.Level) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	slog.SetDefault(slog.New(handler))
}

// WithRequestID はコンテキストにリクエストIDを設定する
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFrom はコンテキストからリクエストIDを取得する
func RequestIDFrom(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// FromContext はコンテキストからリクエストIDを含むロガーを返す
func FromContext(ctx context.Context) *slog.Logger {
	requestID := RequestIDFrom(ctx)
	if requestID != "" {
		return slog.With("request_id", requestID)
	}
	return slog.Default()
}
