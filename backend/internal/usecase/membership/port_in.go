package membership

import "context"

type MembershipUseCase interface {
	CreateMembership(ctx context.Context, cmd CreateMembershipCommand) (*MembershipDTO, error)
	GetMembership(ctx context.Context, query GetMembershipQuery) (*MembershipDTO, error)
	ListMemberships(ctx context.Context, query ListMembershipQuery) (*MembershipSearchResult, error)
	ListByIdolID(ctx context.Context, idolID string) ([]*MembershipDTO, error)
	ListByGroupID(ctx context.Context, groupID string) ([]*MembershipDTO, error)
	UpdateMembership(ctx context.Context, cmd UpdateMembershipCommand) error
	DeleteMembership(ctx context.Context, cmd DeleteMembershipCommand) error
}
