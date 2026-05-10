package usage

import "context"

// Repository は月次使用量リポジトリのインターフェース
type Repository interface {
	// IncrementAndGet は使用量を1増やし、最新の MonthlyUsage を返す
	// ドキュメントが存在しない場合は新規作成して返す
	IncrementAndGet(ctx context.Context, keyPrefix, yearMonth string, limit int) (*MonthlyUsage, error)

	// Get は使用量を取得する（インクリメントなし）
	// 存在しない場合は count=0 の MonthlyUsage を返す
	Get(ctx context.Context, keyPrefix, yearMonth string, limit int) (*MonthlyUsage, error)
}
