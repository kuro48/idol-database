package release

import "context"

// ReleaseUseCase はリリースのユースケース Input Port
type ReleaseUseCase interface {
	CreateRelease(ctx context.Context, cmd CreateReleaseCommand) (*ReleaseDTO, error)
	GetRelease(ctx context.Context, id string) (*ReleaseDTO, error)
	SearchReleases(ctx context.Context, query ListReleasesQuery) (*SearchResult, error)
	UpdateRelease(ctx context.Context, cmd UpdateReleaseCommand) error
	DeleteRelease(ctx context.Context, cmd DeleteReleaseCommand) error
	RestoreRelease(ctx context.Context, id string) error
	UpdateStreamingLinks(ctx context.Context, cmd UpdateStreamingLinksCommand) error
	UpdateExternalIDs(ctx context.Context, cmd UpdateExternalIDsCommand) error
}
