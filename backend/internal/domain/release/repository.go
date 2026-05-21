package release

import "context"

// Repository はリリース集約のリポジトリインターフェース
type Repository interface {
	Save(ctx context.Context, r *Release) error
	FindByID(ctx context.Context, id ReleaseID) (*Release, error)
	Update(ctx context.Context, r *Release) error
	Delete(ctx context.Context, id ReleaseID) error
	Restore(ctx context.Context, id ReleaseID) error
	Search(ctx context.Context, criteria SearchCriteria) ([]*Release, error)
	Count(ctx context.Context, criteria SearchCriteria) (int64, error)
	FindByExternalID(ctx context.Context, kind ReleaseExternalIDKind, value string) (*Release, error)
	FindByArtistID(ctx context.Context, artistID string) ([]*Release, error)
}
