package removal

import "context"

// RemovalUseCase は削除申請のユースケース Input Port
type RemovalUseCase interface {
	CreateRemovalRequest(ctx context.Context, cmd CreateRemovalRequestCommand) (*RemovalRequestDTO, error)
	GetRemovalRequest(ctx context.Context, id string) (*RemovalRequestDTO, error)
	ListAllRemovalRequests(ctx context.Context) ([]*RemovalRequestDTO, error)
	ListPendingRemovalRequests(ctx context.Context) ([]*RemovalRequestDTO, error)
	UpdateStatus(ctx context.Context, cmd UpdateStatusCommand) (*RemovalRequestDTO, error)
}
