package edithistory

import "context"

// SearchCriteria は編集履歴の検索条件
type SearchCriteria struct {
	EntityType *EntityType
	EntityID   *string
	Action     *Action
	ChangedBy  *string
	Offset     int
	Limit      int
}

// Repository は編集履歴リポジトリの契約
type Repository interface {
	Save(ctx context.Context, h *EditHistory) error
	FindByID(ctx context.Context, id EditHistoryID) (*EditHistory, error)
	Search(ctx context.Context, criteria SearchCriteria) ([]*EditHistory, error)
	Count(ctx context.Context, criteria SearchCriteria) (int64, error)
}
