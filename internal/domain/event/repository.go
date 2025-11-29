package event

import (
	"context"
	"time"
)

// SearchCriteria はイベント検索条件
type SearchCriteria struct {
	EventType     *EventType
	StartDateFrom *time.Time
	StartDateTo   *time.Time
	VenueID       *string
	PerformerID   *string
	Tags          []string
	Prefecture    *string // 会場の都道府県（将来実装）

	Sort   string
	Order  string
	Offset int
	Limit  int
}

// Repository はイベント集約のリポジトリインターフェース
type Repository interface {
	// Save は新しいイベントを保存する
	Save(ctx context.Context, event *Event) error

	// FindByID はIDでイベントを検索する
	FindByID(ctx context.Context, id EventID) (*Event, error)

	// Search は条件を指定してイベントを検索する
	Search(ctx context.Context, criteria SearchCriteria) ([]*Event, error)

	// Count は検索条件に一致するイベント数を返す
	Count(ctx context.Context, criteria SearchCriteria) (int64, error)

	// Update は既存のイベントを更新する
	Update(ctx context.Context, event *Event) error

	// Delete はイベントを削除する
	Delete(ctx context.Context, id EventID) error

	// FindUpcoming は今後開催されるイベントを取得する
	FindUpcoming(ctx context.Context, limit int) ([]*Event, error)

	// FindByPerformer はパフォーマーIDでイベントを検索する
	FindByPerformer(ctx context.Context, performerID string, limit int) ([]*Event, error)
}
