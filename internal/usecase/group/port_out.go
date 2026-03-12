package group

import (
	"context"

	domain "github.com/kuro48/idol-api/internal/domain/group"
)

// GroupAppPort は group.Usecase が group application サービスに要求する契約
type GroupAppPort interface {
	CreateGroup(ctx context.Context, input GroupCreateInput) (*domain.Group, error)
	GetGroup(ctx context.Context, id string) (*domain.Group, error)
	ListGroup(ctx context.Context) ([]*domain.Group, error)
	UpdateGroup(ctx context.Context, input GroupUpdateInput) error
	DeleteGroup(ctx context.Context, id string) error
}

// GroupCreateInput はグループ作成の入力
type GroupCreateInput struct {
	Name          string
	FormationDate *string
	DisbandDate   *string
}

// GroupUpdateInput はグループ更新の入力
type GroupUpdateInput struct {
	ID            string
	Name          *string
	FormationDate *string
	DisbandDate   *string
}
