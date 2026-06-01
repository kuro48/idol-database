package adapters

import (
	"context"

	appMembership "github.com/kuro48/idol-api/internal/application/membership"
	domainMembership "github.com/kuro48/idol-api/internal/domain/membership"
	ucMembership "github.com/kuro48/idol-api/internal/usecase/membership"
)

// MembershipAppAdapter は appMembership.ApplicationService を ucMembership.MembershipAppPort に適合させる
type MembershipAppAdapter struct {
	svc *appMembership.ApplicationService
}

func NewMembershipAppAdapter(svc *appMembership.ApplicationService) ucMembership.MembershipAppPort {
	return &MembershipAppAdapter{svc: svc}
}

func (a *MembershipAppAdapter) CreateMembership(ctx context.Context, input ucMembership.MembershipCreateInput) (*domainMembership.Membership, error) {
	return a.svc.CreateMembership(ctx, appMembership.CreateInput{
		IdolID:   input.IdolID,
		GroupID:  input.GroupID,
		Role:     input.Role,
		JoinedAt: input.JoinedAt,
	})
}

func (a *MembershipAppAdapter) GetMembership(ctx context.Context, id string) (*domainMembership.Membership, error) {
	return a.svc.GetMembership(ctx, id)
}

func (a *MembershipAppAdapter) ListByIdolID(ctx context.Context, idolID string) ([]*domainMembership.Membership, error) {
	return a.svc.ListByIdolID(ctx, idolID)
}

func (a *MembershipAppAdapter) ListByGroupID(ctx context.Context, groupID string) ([]*domainMembership.Membership, error) {
	return a.svc.ListByGroupID(ctx, groupID)
}

func (a *MembershipAppAdapter) SearchMemberships(ctx context.Context, criteria domainMembership.SearchCriteria) ([]*domainMembership.Membership, error) {
	return a.svc.SearchMemberships(ctx, criteria)
}

func (a *MembershipAppAdapter) CountMemberships(ctx context.Context, criteria domainMembership.SearchCriteria) (int64, error) {
	return a.svc.CountMemberships(ctx, criteria)
}

func (a *MembershipAppAdapter) UpdateMembership(ctx context.Context, input ucMembership.MembershipUpdateInput) error {
	return a.svc.UpdateMembership(ctx, appMembership.UpdateInput{
		ID:       input.ID,
		Role:     input.Role,
		JoinedAt: input.JoinedAt,
		LeftAt:   input.LeftAt,
	})
}

func (a *MembershipAppAdapter) DeleteMembership(ctx context.Context, id string) error {
	return a.svc.DeleteMembership(ctx, id)
}
