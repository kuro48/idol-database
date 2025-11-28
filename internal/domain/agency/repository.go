package agency

import "context"

// Repository は事務所リポジトリインターフェース
type Repository interface {
	Save(ctx context.Context, agency *Agency) error
	FindByID(ctx context.Context, id AgencyID) (*Agency, error)
	FindAll(ctx context.Context) ([]*Agency, error)
	Update(ctx context.Context, agency *Agency) error
	Delete(ctx context.Context, id AgencyID) error
	ExistsByID(ctx context.Context, id AgencyID) (bool, error)
	ExistsByName(ctx context.Context, name AgencyName) (bool, error)
}
