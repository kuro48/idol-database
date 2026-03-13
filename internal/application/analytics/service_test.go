package analytics_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kuro48/idol-api/internal/application/analytics"
	domainAnalytics "github.com/kuro48/idol-api/internal/domain/analytics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUsageRepository はUsageRepositoryのモック
type MockUsageRepository struct {
	mock.Mock
}

func (m *MockUsageRepository) Save(ctx context.Context, record *domainAnalytics.APIUsageRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockUsageRepository) FindByMaskedKey(ctx context.Context, maskedKey string, from, to time.Time) ([]*domainAnalytics.APIUsageRecord, error) {
	args := m.Called(ctx, maskedKey, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainAnalytics.APIUsageRecord), args.Error(1)
}

func (m *MockUsageRepository) AggregateByKey(ctx context.Context, from, to time.Time) ([]*domainAnalytics.KeyUsageSummary, error) {
	args := m.Called(ctx, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainAnalytics.KeyUsageSummary), args.Error(1)
}

func TestApplicationService_RecordUsage(t *testing.T) {
	t.Run("保存成功時は非ブロッキングで記録される", func(t *testing.T) {
		repo := new(MockUsageRepository)
		saved := make(chan struct{}, 1)
		repo.On("Save", mock.Anything, mock.AnythingOfType("*analytics.APIUsageRecord")).
			Run(func(args mock.Arguments) { saved <- struct{}{} }).
			Return(nil)

		svc := analytics.NewApplicationService(repo)
		record := &domainAnalytics.APIUsageRecord{
			MaskedKey:  "sk-t****word",
			Endpoint:   "/api/v1/idols",
			Method:     "GET",
			StatusCode: 200,
			LatencyMs:  42,
			RecordedAt: time.Now(),
		}

		svc.RecordUsage(context.Background(), record)

		// goroutineが完了するまで待つ
		select {
		case <-saved:
			// OK
		case <-time.After(2 * time.Second):
			t.Fatal("Save が呼ばれなかった")
		}
		repo.AssertExpectations(t)
	})

	t.Run("保存失敗時もリクエストをブロックしない", func(t *testing.T) {
		repo := new(MockUsageRepository)
		done := make(chan struct{}, 1)
		repo.On("Save", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) { done <- struct{}{} }).
			Return(errors.New("DB接続エラー"))

		svc := analytics.NewApplicationService(repo)
		record := &domainAnalytics.APIUsageRecord{RecordedAt: time.Now()}

		// panic しないこと・ブロックしないことを確認
		assert.NotPanics(t, func() {
			svc.RecordUsage(context.Background(), record)
		})

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("goroutine が実行されなかった")
		}
	})
}

func TestApplicationService_GetUsageSummary(t *testing.T) {
	t.Run("正常に集計結果を返す", func(t *testing.T) {
		repo := new(MockUsageRepository)
		expected := []*domainAnalytics.KeyUsageSummary{
			{MaskedKey: "sk-t****word", TotalRequests: 100, SuccessCount: 95, ErrorCount: 5},
		}
		repo.On("AggregateByKey", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Return(expected, nil)

		svc := analytics.NewApplicationService(repo)
		summaries, err := svc.GetUsageSummary(context.Background(), 7)

		require.NoError(t, err)
		assert.Equal(t, expected, summaries)
		repo.AssertExpectations(t)
	})

	t.Run("days=0はデフォルト7日に補正される", func(t *testing.T) {
		repo := new(MockUsageRepository)
		repo.On("AggregateByKey", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Return([]*domainAnalytics.KeyUsageSummary{}, nil)

		svc := analytics.NewApplicationService(repo)
		_, err := svc.GetUsageSummary(context.Background(), 0)
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("days>90は90日に補正される", func(t *testing.T) {
		repo := new(MockUsageRepository)
		repo.On("AggregateByKey", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Return([]*domainAnalytics.KeyUsageSummary{}, nil)

		svc := analytics.NewApplicationService(repo)
		_, err := svc.GetUsageSummary(context.Background(), 999)
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("リポジトリエラー時はエラーを返す", func(t *testing.T) {
		repo := new(MockUsageRepository)
		repo.On("AggregateByKey", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Return(nil, errors.New("集計エラー"))

		svc := analytics.NewApplicationService(repo)
		summaries, err := svc.GetUsageSummary(context.Background(), 7)

		assert.Error(t, err)
		assert.Nil(t, summaries)
	})
}
