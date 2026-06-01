package venue

import "context"

// SearchCriteria は会場検索の条件を表す値オブジェクト
type SearchCriteria struct {
	Name       *string
	Prefecture *string
	Offset     int
	Limit      int
	Sort       string
	Order      string
}

// Repository は会場の永続化操作を定義するインターフェース
type Repository interface {
	Save(ctx context.Context, v *Venue) error
	FindByID(ctx context.Context, id VenueID) (*Venue, error)
	Search(ctx context.Context, criteria SearchCriteria) ([]*Venue, error)
	Count(ctx context.Context, criteria SearchCriteria) (int64, error)
	Update(ctx context.Context, v *Venue) error
	Delete(ctx context.Context, id VenueID) error
}
