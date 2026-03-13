package idol

import (
	"context"

	agencyDomain "github.com/kuro48/idol-api/internal/domain/agency"
	domain "github.com/kuro48/idol-api/internal/domain/idol"
)

// IdolAppPort は idol.Usecase が idol application サービスに要求する契約
type IdolAppPort interface {
	CreateIdol(ctx context.Context, input IdolCreateInput) (*domain.Idol, error)
	GetIdol(ctx context.Context, id string) (*domain.Idol, error)
	ListIdols(ctx context.Context) ([]*domain.Idol, error)
	UpdateIdol(ctx context.Context, input IdolUpdateInput) error
	DeleteIdol(ctx context.Context, id string) error
	RestoreIdol(ctx context.Context, id string) error
	UpdateSocialLinks(ctx context.Context, input IdolUpdateSocialLinksInput) error
	SearchIdols(ctx context.Context, criteria domain.SearchCriteria) ([]*domain.Idol, int64, error)
	FindDuplicateCandidates(ctx context.Context, id string) ([]*domain.DuplicateCandidate, error)
	UpdateExternalIDs(ctx context.Context, input IdolUpdateExternalIDsInput) error
}

// AgencyAppPort は idol.Usecase が agency application サービスに要求する契約
type AgencyAppPort interface {
	GetAgency(ctx context.Context, id string) (*agencyDomain.Agency, error)
}

// IdolCreateInput はアイドル作成の入力
type IdolCreateInput struct {
	Name      string
	Birthdate *string
	AgencyID  *string
	Aliases   []string
}

// IdolUpdateInput はアイドル更新の入力
type IdolUpdateInput struct {
	ID        string
	Name      *string
	Birthdate *string
	AgencyID  *string
	Aliases   []string
}

// IdolUpdateSocialLinksInput はSNS/外部リンク更新の入力
type IdolUpdateSocialLinksInput struct {
	ID              string
	Twitter         *string
	Instagram       *string
	TikTok          *string
	YouTube         *string
	Facebook        *string
	OfficialWebsite *string
	FanClub         *string
}

// IdolUpdateExternalIDsInput は外部IDマッピング更新の入力
type IdolUpdateExternalIDsInput struct {
	ID          string
	ExternalIDs map[string]string
}
