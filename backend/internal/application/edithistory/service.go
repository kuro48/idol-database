package edithistory

import (
	"context"
	"fmt"

	"github.com/kuro48/idol-api/internal/domain/edithistory"
	sharedid "github.com/kuro48/idol-api/internal/shared/id"
)

// ApplicationService は編集履歴アプリケーションサービス
type ApplicationService struct {
	repository edithistory.Repository
}

// NewApplicationService はアプリケーションサービスを作成する
func NewApplicationService(repository edithistory.Repository) *ApplicationService {
	return &ApplicationService{repository: repository}
}

// Record は編集履歴を記録する
func (s *ApplicationService) Record(ctx context.Context, input RecordInput) error {
	entityType, err := edithistory.NewEntityType(input.EntityType)
	if err != nil {
		return fmt.Errorf("エンティティ種別エラー: %w", err)
	}

	action, err := edithistory.NewAction(input.Action)
	if err != nil {
		return fmt.Errorf("アクションエラー: %w", err)
	}

	changes := make(map[string]edithistory.FieldChange, len(input.Changes))
	for field, fc := range input.Changes {
		changes[field] = edithistory.FieldChange{
			Before: fc.Before,
			After:  fc.After,
		}
	}

	h := edithistory.NewEditHistory(entityType, input.EntityID, action, changes, input.ChangedBy)

	id, err := edithistory.NewEditHistoryID(sharedid.Generate())
	if err != nil {
		return fmt.Errorf("ID生成エラー: %w", err)
	}
	h.SetID(id)

	if err := s.repository.Save(ctx, h); err != nil {
		return fmt.Errorf("編集履歴の保存エラー: %w", err)
	}
	return nil
}

// GetHistory は編集履歴を取得する
func (s *ApplicationService) GetHistory(ctx context.Context, id string) (*edithistory.EditHistory, error) {
	hID, err := edithistory.NewEditHistoryID(id)
	if err != nil {
		return nil, fmt.Errorf("IDエラー: %w", err)
	}
	h, err := s.repository.FindByID(ctx, hID)
	if err != nil {
		return nil, fmt.Errorf("編集履歴の取得エラー: %w", err)
	}
	return h, nil
}

// SearchHistory は条件を指定して編集履歴を検索する
func (s *ApplicationService) SearchHistory(ctx context.Context, criteria edithistory.SearchCriteria) ([]*edithistory.EditHistory, int64, error) {
	entries, err := s.repository.Search(ctx, criteria)
	if err != nil {
		return nil, 0, fmt.Errorf("検索エラー: %w", err)
	}
	total, err := s.repository.Count(ctx, criteria)
	if err != nil {
		return nil, 0, fmt.Errorf("件数取得エラー: %w", err)
	}
	return entries, total, nil
}
