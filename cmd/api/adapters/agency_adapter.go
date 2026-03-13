package adapters

import (
	"context"

	appAgency "github.com/kuro48/idol-api/internal/application/agency"
	agencyDomain "github.com/kuro48/idol-api/internal/domain/agency"
	ucAgency "github.com/kuro48/idol-api/internal/usecase/agency"
)

// AgencyAppAdapterForUsecase は appAgency.ApplicationService を ucAgency.AgencyAppPort に適合させる
// （idol usecase 向けの AgencyAppAdapter と区別するため名前を分ける）
type AgencyAppAdapterForUsecase struct {
	svc *appAgency.ApplicationService
}

// NewAgencyAppAdapterForUsecase は AgencyAppAdapterForUsecase を生成する
func NewAgencyAppAdapterForUsecase(svc *appAgency.ApplicationService) ucAgency.AgencyAppPort {
	return &AgencyAppAdapterForUsecase{svc: svc}
}

func (a *AgencyAppAdapterForUsecase) CreateAgency(ctx context.Context, input ucAgency.AgencyCreateInput) (*agencyDomain.Agency, error) {
	return a.svc.CreateAgency(ctx, appAgency.CreateInput{
		Name:            input.Name,
		NameEn:          input.NameEn,
		FoundedDate:     input.FoundedDate,
		Country:         input.Country,
		OfficialWebsite: input.OfficialWebsite,
		Description:     input.Description,
		LogoURL:         input.LogoURL,
	})
}

func (a *AgencyAppAdapterForUsecase) GetAgency(ctx context.Context, id string) (*agencyDomain.Agency, error) {
	return a.svc.GetAgency(ctx, id)
}

func (a *AgencyAppAdapterForUsecase) ListAgencies(ctx context.Context) ([]*agencyDomain.Agency, error) {
	return a.svc.ListAgencies(ctx)
}

func (a *AgencyAppAdapterForUsecase) ListAgenciesWithPagination(ctx context.Context, opts agencyDomain.SearchOptions) (*agencyDomain.SearchResult, error) {
	return a.svc.ListAgenciesWithPagination(ctx, opts)
}

func (a *AgencyAppAdapterForUsecase) UpdateAgency(ctx context.Context, input ucAgency.AgencyUpdateInput) error {
	return a.svc.UpdateAgency(ctx, appAgency.UpdateInput{
		ID:              input.ID,
		Name:            input.Name,
		NameEn:          input.NameEn,
		FoundedDate:     input.FoundedDate,
		OfficialWebsite: input.OfficialWebsite,
		Description:     input.Description,
		LogoURL:         input.LogoURL,
	})
}

func (a *AgencyAppAdapterForUsecase) DeleteAgency(ctx context.Context, id string) error {
	return a.svc.DeleteAgency(ctx, id)
}
