package tag

import (
	"context"
	"fmt"

	"github.com/kuro48/idol-api/internal/domain/tag"
)

// ApplicationService はタグのアプリケーションサービス
type ApplicationService struct {
	repository tag.Repository
}

// NewApplicationService はタグアプリケーションサービスを作成する
func NewApplicationService(repository tag.Repository) *ApplicationService {
	return &ApplicationService{
		repository: repository,
	}
}

// CreateTag はタグを作成する
func (s *ApplicationService) CreateTag(ctx context.Context, input CreateInput) (*tag.Tag, error) {
	// 同名のタグが既に存在しないかチェック
	existing, _ := s.repository.FindByName(ctx, input.Name)
	if existing != nil {
		return nil, fmt.Errorf("タグ名 '%s' は既に使用されています", input.Name)
	}

	// ドメインモデル作成
	t, err := tag.NewTag(input.Name, input.Category, input.Description)
	if err != nil {
		return nil, fmt.Errorf("タグ作成エラー: %w", err)
	}

	// 保存
	if err := s.repository.Save(ctx, t); err != nil {
		return nil, fmt.Errorf("タグ保存エラー: %w", err)
	}

	return t, nil
}

// UpdateTag はタグを更新する
func (s *ApplicationService) UpdateTag(ctx context.Context, input UpdateInput) error {
	// タグIDの検証
	tagID, err := tag.NewTagID(input.ID)
	if err != nil {
		return fmt.Errorf("タグID検証エラー: %w", err)
	}

	// 既存タグの取得
	t, err := s.repository.FindByID(ctx, tagID)
	if err != nil {
		return fmt.Errorf("タグ取得エラー: %w", err)
	}

	// 更新
	if err := t.UpdateName(input.Name); err != nil {
		return fmt.Errorf("タグ名更新エラー: %w", err)
	}
	if err := t.UpdateCategory(input.Category); err != nil {
		return fmt.Errorf("カテゴリ更新エラー: %w", err)
	}
	if err := t.UpdateDescription(input.Description); err != nil {
		return fmt.Errorf("説明更新エラー: %w", err)
	}

	// 保存
	if err := s.repository.Update(ctx, t); err != nil {
		return fmt.Errorf("タグ更新保存エラー: %w", err)
	}

	return nil
}

// DeleteTag はタグを削除する
func (s *ApplicationService) DeleteTag(ctx context.Context, id string) error {
	tagID, err := tag.NewTagID(id)
	if err != nil {
		return fmt.Errorf("タグID検証エラー: %w", err)
	}

	if err := s.repository.Delete(ctx, tagID); err != nil {
		return fmt.Errorf("タグ削除エラー: %w", err)
	}

	return nil
}

// GetTag はタグを取得する
func (s *ApplicationService) GetTag(ctx context.Context, id string) (*tag.Tag, error) {
	tagID, err := tag.NewTagID(id)
	if err != nil {
		return nil, fmt.Errorf("タグID検証エラー: %w", err)
	}

	t, err := s.repository.FindByID(ctx, tagID)
	if err != nil {
		return nil, fmt.Errorf("タグ取得エラー: %w", err)
	}

	return t, nil
}

// SearchTags はタグを検索する
func (s *ApplicationService) SearchTags(ctx context.Context, criteria tag.SearchCriteria) ([]*tag.Tag, int64, error) {
	tags, total, err := s.repository.Search(ctx, criteria)
	if err != nil {
		return nil, 0, fmt.Errorf("タグ検索エラー: %w", err)
	}

	return tags, total, nil
}
