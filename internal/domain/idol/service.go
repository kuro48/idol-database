package idol

import (
	"context"
	"errors"
)

// DomainService はアイドルドメインのドメインサービス
type DomainService struct {
	repository Repository
}

// NewDomainService はドメインサービスを作成する
func NewDomainService(repository Repository) *DomainService {
	return &DomainService{
		repository: repository,
	}
}

// CanCreate はアイドルを作成可能かを判定する
func (s *DomainService) CanCreate(ctx context.Context, name IdolName) error {
	exists, err := s.repository.ExistsByName(ctx, name)
	if err != nil {
		return err
	}

	if exists {
		return errors.New("同じ名前のアイドルが既に存在します")
	}

	return nil
}

// IsDuplicateName は名前の重複をチェックする
func (s *DomainService) IsDuplicateName(ctx context.Context, name IdolName, excludeID *IdolID) (bool, error) {
	idols, err := s.repository.FindAll(ctx)
	if err != nil {
		return false, err
	}

	for _, idol := range idols {
		// 除外するIDがある場合はスキップ
		if excludeID != nil && idol.ID().Equals(*excludeID) {
			continue
		}

		if idol.Name().Value() == name.Value() {
			return true, nil
		}
	}

	return false, nil
}
