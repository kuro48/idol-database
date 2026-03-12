package adapters

import (
	"context"

	appGroup "github.com/kuro48/idol-api/internal/application/group"
	groupDomain "github.com/kuro48/idol-api/internal/domain/group"
	ucGroup "github.com/kuro48/idol-api/internal/usecase/group"
)

// GroupAppAdapter は appGroup.ApplicationService を ucGroup.GroupAppPort に適合させる
type GroupAppAdapter struct {
	svc *appGroup.ApplicationService
}

// NewGroupAppAdapter は GroupAppAdapter を生成する
func NewGroupAppAdapter(svc *appGroup.ApplicationService) ucGroup.GroupAppPort {
	return &GroupAppAdapter{svc: svc}
}

func (a *GroupAppAdapter) CreateGroup(ctx context.Context, input ucGroup.GroupCreateInput) (*groupDomain.Group, error) {
	return a.svc.CreateGroup(ctx, appGroup.CreateInput{
		Name:          input.Name,
		FormationDate: input.FormationDate,
		DisbandDate:   input.DisbandDate,
	})
}

func (a *GroupAppAdapter) GetGroup(ctx context.Context, id string) (*groupDomain.Group, error) {
	return a.svc.GetGroup(ctx, id)
}

func (a *GroupAppAdapter) ListGroup(ctx context.Context) ([]*groupDomain.Group, error) {
	return a.svc.ListGroup(ctx)
}

func (a *GroupAppAdapter) UpdateGroup(ctx context.Context, input ucGroup.GroupUpdateInput) error {
	return a.svc.UpdateGroup(ctx, appGroup.UpdateInput{
		ID:            input.ID,
		Name:          input.Name,
		FormationDate: input.FormationDate,
		DisbandDate:   input.DisbandDate,
	})
}

func (a *GroupAppAdapter) DeleteGroup(ctx context.Context, id string) error {
	return a.svc.DeleteGroup(ctx, id)
}
