package agency

import (
	"context"
	"errors"
)

// DomainService は事務所ドメインサービス
type DomainService struct {
	repository Repository
}

// NewDomainService はドメインサービスを作成する
func NewDomainService(repository Repository) *DomainService {
	return &DomainService{
		repository: repository,
	}
}

// CanCreate は事務所を作成可能かチェックする
func (s *DomainService) CanCreate(ctx context.Context, name AgencyName) error {
	exists, err := s.repository.ExistsByName(ctx, name)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("同じ名前の事務所が既に存在します")
	}
	return nil
}

// IsDuplicateName は名前が重複しているかチェックする（自分自身は除外）
func (s *DomainService) IsDuplicateName(ctx context.Context, name AgencyName, excludeID *AgencyID) (bool, error) {
	exists, err := s.repository.ExistsByName(ctx, name)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	// 自分自身の名前変更の場合は重複とみなさない
	if excludeID != nil {
		existing, err := s.repository.FindByID(ctx, *excludeID)
		if err != nil {
			return false, err
		}
		if existing.Name().Value() == name.Value() {
			return false, nil
		}
	}

	return true, nil
}
