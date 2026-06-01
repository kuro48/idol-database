package membership

import "context"

type SearchCriteria struct {
	IdolID   *string
	GroupID  *string
	IsActive *bool
	Role     *Role
	Offset   int
	Limit    int
	Sort     string
	Order    string
}

type Repository interface {
	Save(ctx context.Context, m *Membership) error
	FindByID(ctx context.Context, id MembershipID) (*Membership, error)
	FindByIdolID(ctx context.Context, idolID string) ([]*Membership, error)
	FindByGroupID(ctx context.Context, groupID string) ([]*Membership, error)
	Search(ctx context.Context, criteria SearchCriteria) ([]*Membership, error)
	Count(ctx context.Context, criteria SearchCriteria) (int64, error)
	Update(ctx context.Context, m *Membership) error
	Delete(ctx context.Context, id MembershipID) error
	Restore(ctx context.Context, id MembershipID) error
}
