// Package adapters は application サービスを usecase の output port interface に適合させるアダプターを提供する。
// usecase 層が application 層を直接 import しないよう、依存の方向を逆転させる。
package adapters

import (
	"context"

	appAgency "github.com/kuro48/idol-api/internal/application/agency"
	appIdol "github.com/kuro48/idol-api/internal/application/idol"
	agencyDomain "github.com/kuro48/idol-api/internal/domain/agency"
	idolDomain "github.com/kuro48/idol-api/internal/domain/idol"
	ucIdol "github.com/kuro48/idol-api/internal/usecase/idol"
)

// IdolAppAdapter は appIdol.ApplicationService を ucIdol.IdolAppPort に適合させる
type IdolAppAdapter struct {
	svc *appIdol.ApplicationService
}

// NewIdolAppAdapter は IdolAppAdapter を生成する
func NewIdolAppAdapter(svc *appIdol.ApplicationService) ucIdol.IdolAppPort {
	return &IdolAppAdapter{svc: svc}
}

func (a *IdolAppAdapter) CreateIdol(ctx context.Context, input ucIdol.IdolCreateInput) (*idolDomain.Idol, error) {
	return a.svc.CreateIdol(ctx, appIdol.CreateInput{
		Name:      input.Name,
		Birthdate: input.Birthdate,
		AgencyID:  input.AgencyID,
		Aliases:   input.Aliases,
	})
}

func (a *IdolAppAdapter) GetIdol(ctx context.Context, id string) (*idolDomain.Idol, error) {
	return a.svc.GetIdol(ctx, id)
}

func (a *IdolAppAdapter) ListIdols(ctx context.Context) ([]*idolDomain.Idol, error) {
	return a.svc.ListIdols(ctx)
}

func (a *IdolAppAdapter) UpdateIdol(ctx context.Context, input ucIdol.IdolUpdateInput) error {
	return a.svc.UpdateIdol(ctx, appIdol.UpdateInput{
		ID:        input.ID,
		Name:      input.Name,
		Birthdate: input.Birthdate,
		AgencyID:  input.AgencyID,
		Aliases:   input.Aliases,
	})
}

func (a *IdolAppAdapter) DeleteIdol(ctx context.Context, id string) error {
	return a.svc.DeleteIdol(ctx, id)
}

func (a *IdolAppAdapter) RestoreIdol(ctx context.Context, id string) error {
	return a.svc.RestoreIdol(ctx, id)
}

func (a *IdolAppAdapter) UpdateSocialLinks(ctx context.Context, input ucIdol.IdolUpdateSocialLinksInput) error {
	return a.svc.UpdateSocialLinks(ctx, appIdol.UpdateSocialLinksInput{
		ID:              input.ID,
		Twitter:         input.Twitter,
		Instagram:       input.Instagram,
		TikTok:          input.TikTok,
		YouTube:         input.YouTube,
		Facebook:        input.Facebook,
		OfficialWebsite: input.OfficialWebsite,
		FanClub:         input.FanClub,
	})
}

func (a *IdolAppAdapter) SearchIdols(ctx context.Context, criteria idolDomain.SearchCriteria) ([]*idolDomain.Idol, int64, error) {
	return a.svc.SearchIdols(ctx, criteria)
}

func (a *IdolAppAdapter) FindDuplicateCandidates(ctx context.Context, id string) ([]*idolDomain.DuplicateCandidate, error) {
	return a.svc.FindDuplicateCandidates(ctx, id)
}

func (a *IdolAppAdapter) UpdateExternalIDs(ctx context.Context, input ucIdol.IdolUpdateExternalIDsInput) error {
	return a.svc.UpdateExternalIDs(ctx, appIdol.UpdateExternalIDsInput{
		ID:          input.ID,
		ExternalIDs: input.ExternalIDs,
	})
}

// AgencyAppAdapter は appAgency.ApplicationService を ucIdol.AgencyAppPort に適合させる
type AgencyAppAdapter struct {
	svc *appAgency.ApplicationService
}

// NewAgencyAppAdapter は AgencyAppAdapter を生成する
func NewAgencyAppAdapter(svc *appAgency.ApplicationService) ucIdol.AgencyAppPort {
	return &AgencyAppAdapter{svc: svc}
}

func (a *AgencyAppAdapter) GetAgency(ctx context.Context, id string) (*agencyDomain.Agency, error) {
	return a.svc.GetAgency(ctx, id)
}
