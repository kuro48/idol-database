package analytics

import "time"

// APIUsageRecord はAPIリクエストの利用記録
type APIUsageRecord struct {
	ID         string
	MaskedKey  string // 表示用マスク済みキー（例: sk-t****word）
	Endpoint   string
	Method     string
	StatusCode int
	LatencyMs  int64
	RecordedAt time.Time
}
