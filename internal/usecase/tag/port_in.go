package tag

import "context"

// TagUseCase はタグのユースケース Input Port
type TagUseCase interface {
	CreateTag(ctx context.Context, cmd CreateTagCommand) (TagDTO, error)
	GetTag(ctx context.Context, id string) (TagDTO, error)
	UpdateTag(ctx context.Context, cmd UpdateTagCommand) error
	DeleteTag(ctx context.Context, id string) error
	SearchTags(ctx context.Context, query SearchQuery, baseURL string) (SearchResult, error)
}
