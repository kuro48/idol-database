package removal

import (
	"context"

	groupDomain "github.com/kuro48/idol-api/internal/domain/group"
	idolDomain "github.com/kuro48/idol-api/internal/domain/idol"
	domain "github.com/kuro48/idol-api/internal/domain/removal"
)

// RemovalAppPort は removal.Usecase が removal application サービスに要求する契約
type RemovalAppPort interface {
	CreateRemovalRequest(ctx context.Context, input RemovalCreateInput) (*domain.RemovalRequest, error)
	GetRemovalRequest(ctx context.Context, id string) (*domain.RemovalRequest, error)
	ListAllRemovalRequests(ctx context.Context) ([]*domain.RemovalRequest, error)
	ListPendingRemovalRequests(ctx context.Context) ([]*domain.RemovalRequest, error)
	UpdateRemovalRequest(ctx context.Context, request *domain.RemovalRequest) error
}

// RemovalIdolPort は removal.Usecase が idol application サービスに要求する契約
type RemovalIdolPort interface {
	GetIdol(ctx context.Context, id string) (*idolDomain.Idol, error)
	DeleteIdol(ctx context.Context, id string) error
}

// RemovalGroupPort は removal.Usecase が group application サービスに要求する契約
type RemovalGroupPort interface {
	GetGroup(ctx context.Context, id string) (*groupDomain.Group, error)
	DeleteGroup(ctx context.Context, id string) error
}

// RemovalCreateInput は削除申請作成の入力
type RemovalCreateInput struct {
	TargetType  string
	TargetID    string
	Requester   string
	Reason      string
	ContactInfo string
	Evidence    string
	Description string
}
