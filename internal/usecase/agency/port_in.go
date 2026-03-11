package agency

import "context"

// AgencyUseCase は事務所のユースケース Input Port
type AgencyUseCase interface {
	CreateAgency(ctx context.Context, cmd CreateAgencyCommand) (*AgencyDTO, error)
	GetAgency(ctx context.Context, query GetAgencyQuery) (*AgencyDTO, error)
	ListAgencies(ctx context.Context, query ListAgenciesQuery) ([]*AgencyDTO, error)
	UpdateAgency(ctx context.Context, cmd UpdateAgencyCommand) error
	DeleteAgency(ctx context.Context, cmd DeleteAgencyCommand) error
}
