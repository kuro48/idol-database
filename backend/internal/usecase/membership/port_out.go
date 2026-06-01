package membership

import (
	"context"

	domain "github.com/kuro48/idol-api/internal/domain/membership"
)

// MembershipAppPort は Usecase が application サービスに要求する契約
type MembershipAppPort interface {
	CreateMembership(ctx context.Context, input MembershipCreateInput) (*domain.Membership, error)
	GetMembership(ctx context.Context, id string) (*domain.Membership, error)
	ListByIdolID(ctx context.Context, idolID string) ([]*domain.Membership, error)
	ListByGroupID(ctx context.Context, groupID string) ([]*domain.Membership, error)
	SearchMemberships(ctx context.Context, criteria domain.SearchCriteria) ([]*domain.Membership, error)
	CountMemberships(ctx context.Context, criteria domain.SearchCriteria) (int64, error)
	UpdateMembership(ctx context.Context, input MembershipUpdateInput) error
	DeleteMembership(ctx context.Context, id string) error
}

type MembershipCreateInput struct {
	IdolID   string
	GroupID  string
	Role     string
	JoinedAt *string
}

type MembershipUpdateInput struct {
	ID       string
	Role     *string
	JoinedAt *string
	LeftAt   *string
}
