package release

import (
	"context"

	appRelease "github.com/kuro48/idol-api/internal/application/release"
	domainRelease "github.com/kuro48/idol-api/internal/domain/release"
)

// ReleaseAppPort は release.Usecase が release application サービスに要求する契約
type ReleaseAppPort interface {
	CreateRelease(ctx context.Context, input appRelease.CreateInput) (*domainRelease.Release, error)
	GetRelease(ctx context.Context, id string) (*domainRelease.Release, error)
	UpdateRelease(ctx context.Context, input appRelease.UpdateInput) error
	DeleteRelease(ctx context.Context, id string) error
	RestoreRelease(ctx context.Context, id string) error
	SearchReleases(ctx context.Context, criteria domainRelease.SearchCriteria) ([]*domainRelease.Release, int64, error)
	UpdateStreamingLinks(ctx context.Context, input appRelease.UpdateStreamingLinksInput) error
	UpdateExternalIDs(ctx context.Context, input appRelease.UpdateExternalIDsInput) error
}

// IdolExistencePort はアイドルの存在確認に使用する
type IdolExistencePort interface {
	GetIdol(ctx context.Context, id string) error
}

// GroupExistencePort はグループの存在確認に使用する
type GroupExistencePort interface {
	GetGroup(ctx context.Context, id string) error
}
