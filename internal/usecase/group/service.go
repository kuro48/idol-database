package group

import (
	"context"

	app "github.com/kuro48/idol-api/internal/application/group"
	domain "github.com/kuro48/idol-api/internal/domain/group"
)

// Usecase はグループのユースケース
type Usecase struct {
	appService *app.ApplicationService
}

// NewUsecase はユースケースを作成する
func NewUsecase(appService *app.ApplicationService) *Usecase {
	return &Usecase{appService: appService}
}

// CreateGroup はグループを作成する
func (u *Usecase) CreateGroup(ctx context.Context, cmd CreateGroupCommand) (*GroupDTO, error) {
	entity, err := u.appService.CreateGroup(ctx, app.CreateInput{
		Name:          cmd.Name,
		FormationDate: cmd.FormationDate,
		DisbandDate:   cmd.DisbandDate,
	})
	if err != nil {
		return nil, err
	}

	dto := toDTO(entity)
	return &dto, nil
}

// GetGroup はグループを取得する
func (u *Usecase) GetGroup(ctx context.Context, query GetGroupQuery) (*GroupDTO, error) {
	entity, err := u.appService.GetGroup(ctx, query.ID)
	if err != nil {
		return nil, err
	}

	dto := toDTO(entity)
	return &dto, nil
}

// ListGroup はグループ一覧を取得する
func (u *Usecase) ListGroup(ctx context.Context, query ListGroupQuery) ([]*GroupDTO, error) {
	groups, err := u.appService.ListGroup(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]*GroupDTO, 0, len(groups))
	for _, g := range groups {
		dto := toDTO(g)
		dtos = append(dtos, &dto)
	}

	return dtos, nil
}

// UpdateGroup はグループを更新する
func (u *Usecase) UpdateGroup(ctx context.Context, cmd UpdateGroupCommand) error {
	return u.appService.UpdateGroup(ctx, app.UpdateInput{
		ID:            cmd.ID,
		Name:          cmd.Name,
		FormationDate: cmd.FormationDate,
		DisbandDate:   cmd.DisbandDate,
	})
}

// DeleteGroup はグループを削除する
func (u *Usecase) DeleteGroup(ctx context.Context, cmd DeleteGroupCommand) error {
	return u.appService.DeleteGroup(ctx, cmd.ID)
}

func toDTO(g *domain.Group) GroupDTO {
	var formationDate string
	if g.FormationDate() != nil {
		formationDate = g.FormationDate().String()
	}

	var disbandDate string
	if g.DisbandDate() != nil {
		disbandDate = g.DisbandDate().String()
	}

	return GroupDTO{
		ID:            g.ID().Value(),
		Name:          g.Name().Value(),
		FormationDate: formationDate,
		DisbandDate:   disbandDate,
		CreatedAt:     g.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     g.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}
