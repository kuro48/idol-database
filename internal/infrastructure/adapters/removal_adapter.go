package adapters

import (
	"context"

	appGroup "github.com/kuro48/idol-api/internal/application/group"
	appIdol "github.com/kuro48/idol-api/internal/application/idol"
	appRemoval "github.com/kuro48/idol-api/internal/application/removal"
	groupDomain "github.com/kuro48/idol-api/internal/domain/group"
	idolDomain "github.com/kuro48/idol-api/internal/domain/idol"
	removalDomain "github.com/kuro48/idol-api/internal/domain/removal"
	ucRemoval "github.com/kuro48/idol-api/internal/usecase/removal"
)

// RemovalAppAdapter は appRemoval.ApplicationService を ucRemoval.RemovalAppPort に適合させる
type RemovalAppAdapter struct {
	svc *appRemoval.ApplicationService
}

// NewRemovalAppAdapter は RemovalAppAdapter を生成する
func NewRemovalAppAdapter(svc *appRemoval.ApplicationService) ucRemoval.RemovalAppPort {
	return &RemovalAppAdapter{svc: svc}
}

func (a *RemovalAppAdapter) CreateRemovalRequest(ctx context.Context, input ucRemoval.RemovalCreateInput) (*removalDomain.RemovalRequest, error) {
	return a.svc.CreateRemovalRequest(ctx, appRemoval.CreateInput{
		TargetType:  input.TargetType,
		TargetID:    input.TargetID,
		Requester:   input.Requester,
		Reason:      input.Reason,
		ContactInfo: input.ContactInfo,
		Evidence:    input.Evidence,
		Description: input.Description,
	})
}

func (a *RemovalAppAdapter) GetRemovalRequest(ctx context.Context, id string) (*removalDomain.RemovalRequest, error) {
	return a.svc.GetRemovalRequest(ctx, id)
}

func (a *RemovalAppAdapter) ListAllRemovalRequests(ctx context.Context) ([]*removalDomain.RemovalRequest, error) {
	return a.svc.ListAllRemovalRequests(ctx)
}

func (a *RemovalAppAdapter) ListPendingRemovalRequests(ctx context.Context) ([]*removalDomain.RemovalRequest, error) {
	return a.svc.ListPendingRemovalRequests(ctx)
}

func (a *RemovalAppAdapter) UpdateRemovalRequest(ctx context.Context, request *removalDomain.RemovalRequest) error {
	return a.svc.UpdateRemovalRequest(ctx, request)
}

// RemovalIdolAdapter は appIdol.ApplicationService を ucRemoval.RemovalIdolPort に適合させる
type RemovalIdolAdapter struct {
	svc *appIdol.ApplicationService
}

// NewRemovalIdolAdapter は RemovalIdolAdapter を生成する
func NewRemovalIdolAdapter(svc *appIdol.ApplicationService) ucRemoval.RemovalIdolPort {
	return &RemovalIdolAdapter{svc: svc}
}

func (a *RemovalIdolAdapter) GetIdol(ctx context.Context, id string) (*idolDomain.Idol, error) {
	return a.svc.GetIdol(ctx, id)
}

func (a *RemovalIdolAdapter) DeleteIdol(ctx context.Context, id string) error {
	return a.svc.DeleteIdol(ctx, id)
}

// RemovalGroupAdapter は appGroup.ApplicationService を ucRemoval.RemovalGroupPort に適合させる
type RemovalGroupAdapter struct {
	svc *appGroup.ApplicationService
}

// NewRemovalGroupAdapter は RemovalGroupAdapter を生成する
func NewRemovalGroupAdapter(svc *appGroup.ApplicationService) ucRemoval.RemovalGroupPort {
	return &RemovalGroupAdapter{svc: svc}
}

func (a *RemovalGroupAdapter) GetGroup(ctx context.Context, id string) (*groupDomain.Group, error) {
	return a.svc.GetGroup(ctx, id)
}

func (a *RemovalGroupAdapter) DeleteGroup(ctx context.Context, id string) error {
	return a.svc.DeleteGroup(ctx, id)
}
