package group

import (
	"context"
	"fmt"

	"github.com/kuro48/idol-api/internal/domain/group"
)

type ApplicationService struct {
	repository    group.Repository
	domainService *group.DomainService
}

func NewApplicationService(repository group.Repository) *ApplicationService {
	return &ApplicationService{
		repository:    repository,
		domainService: group.NewDomainService(repository),
	}
}

func (s *ApplicationService) CreateGroup(ctx context.Context, input CreateInput) (*group.Group, error) {
	name, err := group.NewGroupName(input.Name)
	if err != nil {
		return nil, fmt.Errorf("名前の生成エラー: %w", err)
	}

	// ドメインサービスで重複チェック
	if err := s.domainService.CanCreate(ctx, name); err != nil {
		return nil, err
	}

	var formationDate *group.FormationDate
	if input.FormationDate != nil {
		fd, err := group.NewFormationDateFromString(*input.FormationDate)
		if err != nil {
			return nil, fmt.Errorf("結成日の生成エラー: %w", err)
		}
		formationDate = &fd
	}

	newGroup, err := group.NewGroup(name, formationDate)
	if err != nil {
		return nil, fmt.Errorf("グループの生成エラー: %w", err)
	}

	if input.DisbandDate != nil {
		d, err := group.NewDisbandDateFromString(*input.DisbandDate)
		if err != nil {
			return nil, fmt.Errorf("解散日の生成エラー: %w", err)
		}
		newGroup.UpdateDisbandDate(d)
	}

	if err := s.repository.Save(ctx, newGroup); err != nil {
		return nil, fmt.Errorf("グループの保存エラー: %w", err)
	}

	return newGroup, nil
}

func (s *ApplicationService) GetGroup(ctx context.Context, id string) (*group.Group, error) {
	groupID, err := group.NewGroupID(id)
	if err != nil {
		return nil, fmt.Errorf("IDの生成エラー: %w", err)
	}

	foundGroup, err := s.repository.FindByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("グループの取得エラー: %w", err)
	}

	return foundGroup, nil
}

func (s *ApplicationService) ListGroup(ctx context.Context) ([]*group.Group, error) {
	groups, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("グループ一覧の取得エラー: %w", err)
	}

	return groups, nil
}

func (s *ApplicationService) UpdateGroup(ctx context.Context, input UpdateInput) error {
	id, err := group.NewGroupID(input.ID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	existingGroup, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("グループの取得エラー: %w", err)
	}

	// 各フィールドの更新
	if input.Name != nil {
		name, err := group.NewGroupName(*input.Name)
		if err != nil {
			return fmt.Errorf("名前の生成エラー: %w", err)
		}

		// 名前の重複チェック（自分自身は除外）
		isDuplicate, err := s.domainService.IsDuplicateName(ctx, name, &id)
		if err != nil {
			return err
		}
		if isDuplicate {
			return fmt.Errorf("同じ名前のグループが既に存在します")
		}

		if err := existingGroup.ChangeName(name); err != nil {
			return err
		}
	}

	if input.FormationDate != nil {
		fd, err := group.NewFormationDateFromString(*input.FormationDate)
		if err != nil {
			return fmt.Errorf("結成日の生成エラー: %w", err)
		}
		existingGroup.UpdateFormationDate(fd)
	}

	if input.DisbandDate != nil {
		d, err := group.NewDisbandDateFromString(*input.DisbandDate)
		if err != nil {
			return fmt.Errorf("解散日の生成エラー: %w", err)
		}
		existingGroup.UpdateDisbandDate(d)
	}

	// 更新の保存
	if err := s.repository.Update(ctx, existingGroup); err != nil {
		return fmt.Errorf("グループの更新エラー: %w", err)
	}

	return nil
}

func (s *ApplicationService) DeleteGroup(ctx context.Context, id string) error {
	groupID, err := group.NewGroupID(id)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	if err := s.repository.Delete(ctx, groupID); err != nil {
		return fmt.Errorf("グループの削除エラー: %w", err)
	}

	return nil
}
