package group

import (
	"context"
	"errors"
)

// DomainService はグループドメインのドメインサービス
type DomainService struct {
	repository Repository
}

// NewDomainService はドメインサービスを作成する
func NewDomainService(repository Repository) *DomainService {
	return &DomainService{
		repository: repository,
	}
}

// CanCreate はグループを作成可能かを判定する
func (s *DomainService) CanCreate(ctx context.Context, name GroupName) error {
	exists, err := s.repository.ExistsByName(ctx, name)
	if err != nil {
		return err
	}

	if exists {
		return errors.New("同じ名前のグループが既に存在します")
	}

	return nil
}

// IsDuplicateName は名前の重複をチェックする
func (s *DomainService) IsDuplicateName(ctx context.Context, name GroupName, excludeID *GroupID) (bool, error) {
	groups, err := s.repository.FindAll(ctx)
	if err != nil {
		return false, err
	}

	for _, group := range groups {
		// 除外するIDがある場合はスキップ
		if excludeID != nil && group.ID().Equals(*excludeID) {
			continue
		}

		if group.Name().Value() == name.Value() {
			return true, nil
		}
	}

	return false, nil
}
