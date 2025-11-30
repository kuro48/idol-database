package tag

import (
	"context"
	"fmt"
	"math"

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
func (s *ApplicationService) CreateTag(ctx context.Context, cmd CreateTagCommand) (TagDTO, error) {
	// 同名のタグが既に存在しないかチェック
	existing, _ := s.repository.FindByName(ctx, cmd.Name)
	if existing != nil {
		return TagDTO{}, fmt.Errorf("タグ名 '%s' は既に使用されています", cmd.Name)
	}

	// ドメインモデル作成
	t, err := tag.NewTag(cmd.Name, cmd.Category, cmd.Description)
	if err != nil {
		return TagDTO{}, fmt.Errorf("タグ作成エラー: %w", err)
	}

	// 保存
	if err := s.repository.Save(ctx, t); err != nil {
		return TagDTO{}, fmt.Errorf("タグ保存エラー: %w", err)
	}

	return ToDTO(t), nil
}

// UpdateTag はタグを更新する
func (s *ApplicationService) UpdateTag(ctx context.Context, cmd UpdateTagCommand) error {
	// タグIDの検証
	tagID, err := tag.NewTagID(cmd.ID)
	if err != nil {
		return fmt.Errorf("タグID検証エラー: %w", err)
	}

	// 既存タグの取得
	t, err := s.repository.FindByID(ctx, tagID)
	if err != nil {
		return fmt.Errorf("タグ取得エラー: %w", err)
	}

	// 更新
	if err := t.UpdateName(cmd.Name); err != nil {
		return fmt.Errorf("タグ名更新エラー: %w", err)
	}
	if err := t.UpdateCategory(cmd.Category); err != nil {
		return fmt.Errorf("カテゴリ更新エラー: %w", err)
	}
	if err := t.UpdateDescription(cmd.Description); err != nil {
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
func (s *ApplicationService) GetTag(ctx context.Context, id string) (TagDTO, error) {
	tagID, err := tag.NewTagID(id)
	if err != nil {
		return TagDTO{}, fmt.Errorf("タグID検証エラー: %w", err)
	}

	t, err := s.repository.FindByID(ctx, tagID)
	if err != nil {
		return TagDTO{}, fmt.Errorf("タグ取得エラー: %w", err)
	}

	return ToDTO(t), nil
}

// SearchTags はタグを検索する
func (s *ApplicationService) SearchTags(ctx context.Context, query SearchQuery, baseURL string) (SearchResult, error) {
	// クエリをドメインの検索条件に変換
	criteria, err := query.ToCriteria()
	if err != nil {
		return SearchResult{}, fmt.Errorf("検索条件変換エラー: %w", err)
	}

	// 検索実行
	tags, total, err := s.repository.Search(ctx, criteria)
	if err != nil {
		return SearchResult{}, fmt.Errorf("タグ検索エラー: %w", err)
	}

	// DTOに変換
	dtos := make([]TagDTO, 0, len(tags))
	for _, t := range tags {
		dtos = append(dtos, ToDTO(t))
	}

	// ページネーション情報を構築
	page := query.Page
	if page < 1 {
		page = 1
	}
	limit := query.Limit
	if limit < 1 {
		limit = 20
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	meta := PaginationMeta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	// ページネーションリンクを構築
	links := s.buildPaginationLinks(baseURL, query, totalPages)

	return SearchResult{
		Data:  dtos,
		Meta:  meta,
		Links: links,
	}, nil
}

// buildPaginationLinks はページネーションリンクを構築する
func (s *ApplicationService) buildPaginationLinks(baseURL string, query SearchQuery, totalPages int) PaginationLinks {
	buildURL := func(page int) string {
		url := fmt.Sprintf("%s?page=%d&limit=%d", baseURL, page, query.Limit)
		if query.Name != nil {
			url += fmt.Sprintf("&name=%s", *query.Name)
		}
		if query.Category != nil {
			url += fmt.Sprintf("&category=%s", *query.Category)
		}
		return url
	}

	links := PaginationLinks{
		First: buildURL(1),
		Last:  buildURL(totalPages),
	}

	if query.Page > 1 {
		links.Prev = buildURL(query.Page - 1)
	}

	if query.Page < totalPages {
		links.Next = buildURL(query.Page + 1)
	}

	return links
}
