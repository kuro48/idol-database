package agency

import (
	"context"

	domain "github.com/kuro48/idol-api/internal/domain/agency"
)

// AgencyAppPort は agency.Usecase が agency application サービスに要求する契約
type AgencyAppPort interface {
	CreateAgency(ctx context.Context, input AgencyCreateInput) (*domain.Agency, error)
	GetAgency(ctx context.Context, id string) (*domain.Agency, error)
	ListAgencies(ctx context.Context) ([]*domain.Agency, error)
	UpdateAgency(ctx context.Context, input AgencyUpdateInput) error
	DeleteAgency(ctx context.Context, id string) error
}

// AgencyCreateInput は事務所作成の入力
type AgencyCreateInput struct {
	Name            string
	NameEn          *string
	FoundedDate     *string
	Country         string
	OfficialWebsite *string
	Description     *string
	LogoURL         *string
}

// AgencyUpdateInput は事務所更新の入力
type AgencyUpdateInput struct {
	ID              string
	Name            *string
	NameEn          *string
	FoundedDate     *string
	OfficialWebsite *string
	Description     *string
	LogoURL         *string
}
