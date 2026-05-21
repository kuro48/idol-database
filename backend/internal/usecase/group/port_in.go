package group

import "context"

// GroupUseCase はグループのユースケース Input Port
type GroupUseCase interface {
	CreateGroup(ctx context.Context, cmd CreateGroupCommand) (*GroupDTO, error)
	GetGroup(ctx context.Context, query GetGroupQuery) (*GroupDTO, error)
	ListGroup(ctx context.Context, query ListGroupQuery) (*GroupSearchResult, error)
	UpdateGroup(ctx context.Context, cmd UpdateGroupCommand) error
	DeleteGroup(ctx context.Context, cmd DeleteGroupCommand) error
}
