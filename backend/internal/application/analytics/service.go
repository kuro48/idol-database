package analytics

import (
	"context"
	"log/slog"
	"time"

	domainAnalytics "github.com/kuro48/idol-api/internal/domain/analytics"
)

// maxConcurrentSaves はAPI利用記録の最大同時保存数
const maxConcurrentSaves = 50

// recordSaveTimeout はAPI利用記録保存のタイムアウト
const recordSaveTimeout = 5 * time.Second

// ApplicationService はAPI利用分析のアプリケーションサービス
type ApplicationService struct {
	repo domainAnalytics.UsageRepository
	sem  chan struct{}
}

// NewApplicationService はアプリケーションサービスを作成する
func NewApplicationService(repo domainAnalytics.UsageRepository) *ApplicationService {
	return &ApplicationService{
		repo: repo,
		sem:  make(chan struct{}, maxConcurrentSaves),
	}
}

// RecordUsage はAPI利用記録を保存する（非ブロッキング、上限超過時はドロップ）
func (s *ApplicationService) RecordUsage(ctx context.Context, record *domainAnalytics.APIUsageRecord) {
	select {
	case s.sem <- struct{}{}:
		go func() {
			defer func() { <-s.sem }()
			saveCtx, cancel := context.WithTimeout(context.Background(), recordSaveTimeout)
			defer cancel()
			if err := s.repo.Save(saveCtx, record); err != nil {
				slog.Error("API利用記録の保存に失敗しました",
					"error", err,
					"endpoint", record.Endpoint,
					"method", record.Method,
				)
			}
		}()
	default:
		slog.Warn("API利用記録をスキップします（同時保存数上限）",
			"endpoint", record.Endpoint,
			"method", record.Method,
		)
	}
}

// GetUsageSummary はAPIキー単位の利用サマリーを取得する
func (s *ApplicationService) GetUsageSummary(ctx context.Context, days int) ([]*domainAnalytics.KeyUsageSummary, error) {
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}

	from := time.Now().AddDate(0, 0, -days)
	to := time.Now()

	summaries, err := s.repo.AggregateByKey(ctx, from, to)
	if err != nil {
		return nil, err
	}

	return summaries, nil
}
