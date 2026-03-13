package analytics

import (
	"context"
	"fmt"
	"time"

	domainAnalytics "github.com/kuro48/idol-api/internal/domain/analytics"
)

// ApplicationService はAPI利用分析のアプリケーションサービス
type ApplicationService struct {
	repo domainAnalytics.UsageRepository
}

// NewApplicationService はアプリケーションサービスを作成する
func NewApplicationService(repo domainAnalytics.UsageRepository) *ApplicationService {
	return &ApplicationService{repo: repo}
}

// RecordUsage はAPI利用記録を保存する（非ブロッキング）
func (s *ApplicationService) RecordUsage(ctx context.Context, record *domainAnalytics.APIUsageRecord) {
	go func() {
		if err := s.repo.Save(context.Background(), record); err != nil {
			// ロギングのみ、利用記録の失敗はリクエストに影響させない
			_ = fmt.Errorf("API利用記録の保存に失敗しました: %w", err)
		}
	}()
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
		return nil, fmt.Errorf("利用統計の取得エラー: %w", err)
	}

	return summaries, nil
}
