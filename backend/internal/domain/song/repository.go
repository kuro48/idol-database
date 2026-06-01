package song

import "context"

// SearchCriteria は楽曲検索条件
type SearchCriteria struct {
	Title  *string
	ISRC   *string
	Offset int
	Limit  int
	Sort   string
	Order  string
}

// Repository は楽曲の永続化インターフェース
type Repository interface {
	Save(ctx context.Context, s *Song) error
	FindByID(ctx context.Context, id SongID) (*Song, error)
	Search(ctx context.Context, criteria SearchCriteria) ([]*Song, error)
	Count(ctx context.Context, criteria SearchCriteria) (int64, error)
	Update(ctx context.Context, s *Song) error
	Delete(ctx context.Context, id SongID) error
	Restore(ctx context.Context, id SongID) error
}
