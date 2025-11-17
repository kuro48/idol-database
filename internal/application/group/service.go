package group

import (
	"context"
	"fmt"

	"github.com/kuro48/idol-api/internal/domain/group"
)

type ApplicationService struct {
	repository group.Repository
	domainService *group.DomainService
}

func NewApplicationService(repository group.Repository) *ApplicationService {
	return &ApplicationService{
		repository: repository,
		domainService: group.NewDomainService(repository),
	}
}

func (s *ApplicationService) CreateGroup(ctx context.Context, cmd CreateGroupCommand) (*GroupDTO, error) {
	name, err := group.NewGroupName(cmd.Name)
	if err != nil {
		return nil, fmt.Errorf("名前の生成エラー: %w", err)
	}

	// ドメインサービスで重複チェック
	if err := s.domainService.CanCreate(ctx, name); err != nil {
		return nil, err
	}

	var formationDate *group.FormationDate
	if cmd.FormationDate != nil {
		fd, err := group.NewFormationDateFromString(*cmd.FormationDate)
		if err != nil {
			return  nil, fmt.Errorf("結成日の生成エラー: %w", err)
		}
		formationDate = &fd
	}

	newGroup, err := group.NewGroup(name, formationDate)
	if err != nil {
		return nil, fmt.Errorf("グループの生成エラー: %w", err)
	}

	if err := s.repository.Save(ctx, newGroup); err != nil {
		return nil, fmt.Errorf("グループの保存エラー: %w", err)
	}

	return s.toDTO(newGroup), nil
}

func (s *ApplicationService) GetGroup(ctx context.Context, query GetGroupQuery) (*GroupDTO, error) {
	id, err := group.NewGroupID(query.ID)
	if err != nil {
		return nil, fmt.Errorf("IDの生成エラー: %w", err)
	}

	foundGroup, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("グループの取得エラー: %w", err)
	}

	return s.toDTO(foundGroup), nil
}

func (s *ApplicationService) ListGroup(ctx context.Context, query ListGroupQuery) ([]*GroupDTO, error) {
	groups, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("グループ一覧の取得エラー: %w", err)
	}

	dtos := make([]*GroupDTO, 0, len(groups))
	for _, g := range groups {
		dtos = append(dtos, s.toDTO(g))
	}

	return dtos, nil
}

func (s *ApplicationService) UpdateGroup(ctx context.Context, cmd UpdateGroupCommand) error {
	id, err := group.NewGroupID(cmd.ID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	existingGroup, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("アイドルの取得エラー: %w", err)
	}

	// 各フィールドの更新
	if cmd.Name != nil {
		name, err := group.NewGroupName(*cmd.Name)
		if err != nil {
			return fmt.Errorf("名前の生成エラー: %w", err)
		}

		// 名前の重複チェック（自分自身は除外）
		isDuplicate, err := s.domainService.IsDuplicateName(ctx, name, &id)
		if err != nil {
			return err
		}
		if isDuplicate {
			return fmt.Errorf("同じ名前のアイドルが既に存在します")
		}

		if err := existingGroup.ChangeName(name); err != nil {
			return err
		}
	}

	if cmd.FormationDate != nil {
		fd, err := group.NewFormationDateFromString(*cmd.FormationDate)
		if err != nil {
			return fmt.Errorf("結成日の生成エラー: %w", err)
		}
		existingGroup.UpdateFormationDate(fd)
	}

		if cmd.DisbandDate != nil {
		fd, err := group.NewDisbandDateFromString(*cmd.DisbandDate)
		if err != nil {
			return fmt.Errorf("解散日の生成エラー: %w", err)
		}
		existingGroup.UpdateDisbandDate(fd)
	}

	// 更新の保存
	if err := s.repository.Update(ctx, existingGroup); err != nil {
		return fmt.Errorf("グループの更新エラー: %w", err)
	}

	return nil
}


func (s *ApplicationService) DeleteGroup(ctx context.Context, cmd DeleteGroupCommand) error {
	id, err := group.NewGroupID(cmd.ID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("グループの削除エラー: %w", err)
	}

	return nil
}
// toDTO はドメインモデルをDTOに変換する
func (s *ApplicationService) toDTO(g *group.Group) *GroupDTO {
	var formationDate string
	if g.FormationDate() != nil {
		formationDate = g.FormationDate().String()
	}

	var disbandDate string
	if g.DisbandDate() != nil {
		disbandDate = g.DisbandDate().String()
	}

	return &GroupDTO{
		ID:            g.ID().Value(),
		Name:          g.Name().Value(),
		FormationDate: formationDate,
		DisbandDate:   disbandDate,
		CreatedAt:     g.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     g.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}
