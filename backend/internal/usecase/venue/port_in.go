package venue

import "context"

// VenueUseCase は会場ユースケースのポート（プレゼンテーション層向け）
type VenueUseCase interface {
	CreateVenue(ctx context.Context, cmd CreateVenueCommand) (*VenueDTO, error)
	GetVenue(ctx context.Context, query GetVenueQuery) (*VenueDTO, error)
	ListVenues(ctx context.Context, query ListVenueQuery) (*VenueSearchResult, error)
	UpdateVenue(ctx context.Context, cmd UpdateVenueCommand) error
	DeleteVenue(ctx context.Context, cmd DeleteVenueCommand) error
}
