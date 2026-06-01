package adapters

import (
	"context"

	appVenue "github.com/kuro48/idol-api/internal/application/venue"
	domainVenue "github.com/kuro48/idol-api/internal/domain/venue"
	ucVenue "github.com/kuro48/idol-api/internal/usecase/venue"
)

type VenueAppAdapter struct {
	svc *appVenue.ApplicationService
}

func NewVenueAppAdapter(svc *appVenue.ApplicationService) ucVenue.VenueAppPort {
	return &VenueAppAdapter{svc: svc}
}

func (a *VenueAppAdapter) CreateVenue(ctx context.Context, input ucVenue.VenueCreateInput) (*domainVenue.Venue, error) {
	return a.svc.CreateVenue(ctx, appVenue.CreateInput{
		Name:        input.Name,
		NameEn:      input.NameEn,
		Prefecture:  input.Prefecture,
		City:        input.City,
		Address:     input.Address,
		Capacity:    input.Capacity,
		OfficialURL: input.OfficialURL,
	})
}

func (a *VenueAppAdapter) GetVenue(ctx context.Context, id string) (*domainVenue.Venue, error) {
	return a.svc.GetVenue(ctx, id)
}

func (a *VenueAppAdapter) SearchVenues(ctx context.Context, criteria domainVenue.SearchCriteria) ([]*domainVenue.Venue, error) {
	return a.svc.SearchVenues(ctx, criteria)
}

func (a *VenueAppAdapter) CountVenues(ctx context.Context, criteria domainVenue.SearchCriteria) (int64, error) {
	return a.svc.CountVenues(ctx, criteria)
}

func (a *VenueAppAdapter) UpdateVenue(ctx context.Context, input ucVenue.VenueUpdateInput) error {
	return a.svc.UpdateVenue(ctx, appVenue.UpdateInput{
		ID:          input.ID,
		Name:        input.Name,
		NameEn:      input.NameEn,
		Prefecture:  input.Prefecture,
		City:        input.City,
		Address:     input.Address,
		Capacity:    input.Capacity,
		OfficialURL: input.OfficialURL,
	})
}

func (a *VenueAppAdapter) DeleteVenue(ctx context.Context, id string) error {
	return a.svc.DeleteVenue(ctx, id)
}
