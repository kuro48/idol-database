package analytics

import (
	"context"
	"time"
)

// KeyUsageSummary はAPIキー単位の利用サマリー
type KeyUsageSummary struct {
	MaskedKey     string
	TotalRequests int64
	SuccessCount  int64
	ErrorCount    int64
	AvgLatencyMs  float64
	LastUsedAt    time.Time
}

// UsageRepository はAPI利用記録のリポジトリインターフェース
type UsageRepository interface {
	Save(ctx context.Context, record *APIUsageRecord) error
	FindByMaskedKey(ctx context.Context, maskedKey string, from, to time.Time) ([]*APIUsageRecord, error)
	AggregateByKey(ctx context.Context, from, to time.Time) ([]*KeyUsageSummary, error)
}
