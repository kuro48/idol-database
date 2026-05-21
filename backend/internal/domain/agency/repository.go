package agency

import "context"

// SearchOptions は検索オプション
type SearchOptions struct {
	Name    *string
	Country *string
	Sort    string
	Order   string
	Page    int
	Limit   int
}

// SearchResult は検索結果
type SearchResult struct {
	Agencies []*Agency
	Total    int64
}

// Repository は事務所リポジトリインターフェース
type Repository interface {
	Save(ctx context.Context, agency *Agency) error
	FindByID(ctx context.Context, id AgencyID) (*Agency, error)
	FindAll(ctx context.Context) ([]*Agency, error)
	FindWithPagination(ctx context.Context, opts SearchOptions) (*SearchResult, error)
	Update(ctx context.Context, agency *Agency) error
	Delete(ctx context.Context, id AgencyID) error
	// Restore はソフトデリートされた事務所を復元する
	Restore(ctx context.Context, id AgencyID) error
	ExistsByID(ctx context.Context, id AgencyID) (bool, error)
	ExistsByName(ctx context.Context, name AgencyName) (bool, error)
}
