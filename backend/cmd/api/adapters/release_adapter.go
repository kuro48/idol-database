package adapters

import (
	"context"

	appGroup "github.com/kuro48/idol-api/internal/application/group"
	appIdol "github.com/kuro48/idol-api/internal/application/idol"
	appRelease "github.com/kuro48/idol-api/internal/application/release"
	domainRelease "github.com/kuro48/idol-api/internal/domain/release"
	ucRelease "github.com/kuro48/idol-api/internal/usecase/release"
)

// ReleaseAppAdapter は appRelease.ApplicationService を ucRelease.ReleaseAppPort に適合させる
type ReleaseAppAdapter struct {
	svc *appRelease.ApplicationService
}

// NewReleaseAppAdapter は ReleaseAppAdapter を生成する
func NewReleaseAppAdapter(svc *appRelease.ApplicationService) ucRelease.ReleaseAppPort {
	return &ReleaseAppAdapter{svc: svc}
}

func (a *ReleaseAppAdapter) CreateRelease(ctx context.Context, input appRelease.CreateInput) (*domainRelease.Release, error) {
	return a.svc.CreateRelease(ctx, input)
}

func (a *ReleaseAppAdapter) GetRelease(ctx context.Context, id string) (*domainRelease.Release, error) {
	return a.svc.GetRelease(ctx, id)
}

func (a *ReleaseAppAdapter) UpdateRelease(ctx context.Context, input appRelease.UpdateInput) error {
	return a.svc.UpdateRelease(ctx, input)
}

func (a *ReleaseAppAdapter) DeleteRelease(ctx context.Context, id string) error {
	return a.svc.DeleteRelease(ctx, id)
}

func (a *ReleaseAppAdapter) RestoreRelease(ctx context.Context, id string) error {
	return a.svc.RestoreRelease(ctx, id)
}

func (a *ReleaseAppAdapter) SearchReleases(ctx context.Context, criteria domainRelease.SearchCriteria) ([]*domainRelease.Release, int64, error) {
	return a.svc.SearchReleases(ctx, criteria)
}

func (a *ReleaseAppAdapter) UpdateStreamingLinks(ctx context.Context, input appRelease.UpdateStreamingLinksInput) error {
	return a.svc.UpdateStreamingLinks(ctx, input)
}

func (a *ReleaseAppAdapter) UpdateExternalIDs(ctx context.Context, input appRelease.UpdateExternalIDsInput) error {
	return a.svc.UpdateExternalIDs(ctx, input)
}

// IdolExistenceAdapter は appIdol.ApplicationService を ucRelease.IdolExistencePort に適合させる
type IdolExistenceAdapter struct {
	svc *appIdol.ApplicationService
}

// NewIdolExistenceAdapter は IdolExistenceAdapter を生成する
func NewIdolExistenceAdapter(svc *appIdol.ApplicationService) ucRelease.IdolExistencePort {
	return &IdolExistenceAdapter{svc: svc}
}

func (a *IdolExistenceAdapter) GetIdol(ctx context.Context, id string) error {
	_, err := a.svc.GetIdol(ctx, id)
	return err
}

// GroupExistenceAdapter は appGroup.ApplicationService を ucRelease.GroupExistencePort に適合させる
type GroupExistenceAdapter struct {
	svc *appGroup.ApplicationService
}

// NewGroupExistenceAdapter は GroupExistenceAdapter を生成する
func NewGroupExistenceAdapter(svc *appGroup.ApplicationService) ucRelease.GroupExistencePort {
	return &GroupExistenceAdapter{svc: svc}
}

func (a *GroupExistenceAdapter) GetGroup(ctx context.Context, id string) error {
	_, err := a.svc.GetGroup(ctx, id)
	return err
}
