package idol

import "context"

// IdolUseCase はアイドルのユースケース Input Port
type IdolUseCase interface {
	CreateIdol(ctx context.Context, cmd CreateIdolCommand) (*IdolDTO, error)
	GetIdol(ctx context.Context, query GetIdolQuery) (*IdolDTO, error)
	ListIdols(ctx context.Context, query ListIdolsQuery) ([]*IdolDTO, error)
	SearchIdols(ctx context.Context, query ListIdolsQuery) (*SearchResult, error)
	UpdateIdol(ctx context.Context, cmd UpdateIdolCommand) error
	DeleteIdol(ctx context.Context, cmd DeleteIdolCommand) error
	UpdateSocialLinks(ctx context.Context, cmd UpdateSocialLinksCommand) error
}
