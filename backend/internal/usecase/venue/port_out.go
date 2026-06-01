package venue

import (
	"context"

	domain "github.com/kuro48/idol-api/internal/domain/venue"
)

// VenueAppPort は Usecase が application サービスに要求する契約
type VenueAppPort interface {
	CreateVenue(ctx context.Context, input VenueCreateInput) (*domain.Venue, error)
	GetVenue(ctx context.Context, id string) (*domain.Venue, error)
	SearchVenues(ctx context.Context, criteria domain.SearchCriteria) ([]*domain.Venue, error)
	CountVenues(ctx context.Context, criteria domain.SearchCriteria) (int64, error)
	UpdateVenue(ctx context.Context, input VenueUpdateInput) error
	DeleteVenue(ctx context.Context, id string) error
}

// VenueCreateInput は会場作成の入力データ（usecase→application）
type VenueCreateInput struct {
	Name        string
	NameEn      *string
	Prefecture  *string
	City        *string
	Address     *string
	Capacity    *int
	OfficialURL *string
}

// VenueUpdateInput は会場更新の入力データ（usecase→application）
type VenueUpdateInput struct {
	ID          string
	Name        *string
	NameEn      *string
	Prefecture  *string
	City        *string
	Address     *string
	Capacity    *int
	OfficialURL *string
}
