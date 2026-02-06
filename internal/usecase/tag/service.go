package tag

import (
	"context"
	"fmt"
	"math"

	app "github.com/kuro48/idol-api/internal/application/tag"
)

// Usecase はタグのユースケース
type Usecase struct {
	appService *app.ApplicationService
}

// NewUsecase はユースケースを作成する
func NewUsecase(appService *app.ApplicationService) *Usecase {
	return &Usecase{appService: appService}
}

// CreateTag はタグを作成する
func (u *Usecase) CreateTag(ctx context.Context, cmd CreateTagCommand) (TagDTO, error) {
	entity, err := u.appService.CreateTag(ctx, app.CreateInput{
		Name:        cmd.Name,
		Category:    cmd.Category,
		Description: cmd.Description,
	})
	if err != nil {
		return TagDTO{}, err
	}

	return ToDTO(entity), nil
}

// UpdateTag はタグを更新する
func (u *Usecase) UpdateTag(ctx context.Context, cmd UpdateTagCommand) error {
	return u.appService.UpdateTag(ctx, app.UpdateInput{
		ID:          cmd.ID,
		Name:        cmd.Name,
		Category:    cmd.Category,
		Description: cmd.Description,
	})
}

// DeleteTag はタグを削除する
func (u *Usecase) DeleteTag(ctx context.Context, id string) error {
	return u.appService.DeleteTag(ctx, id)
}

// GetTag はタグを取得する
func (u *Usecase) GetTag(ctx context.Context, id string) (TagDTO, error) {
	entity, err := u.appService.GetTag(ctx, id)
	if err != nil {
		return TagDTO{}, err
	}

	return ToDTO(entity), nil
}

// SearchTags はタグを検索する
func (u *Usecase) SearchTags(ctx context.Context, query SearchQuery, baseURL string) (SearchResult, error) {
	criteria, err := query.ToCriteria()
	if err != nil {
		return SearchResult{}, fmt.Errorf("検索条件変換エラー: %w", err)
	}

	tags, total, err := u.appService.SearchTags(ctx, criteria)
	if err != nil {
		return SearchResult{}, err
	}

	dtos := make([]TagDTO, 0, len(tags))
	for _, t := range tags {
		dtos = append(dtos, ToDTO(t))
	}

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

	links := buildPaginationLinks(baseURL, query, totalPages)

	return SearchResult{
		Data:  dtos,
		Meta:  meta,
		Links: links,
	}, nil
}

func buildPaginationLinks(baseURL string, query SearchQuery, totalPages int) PaginationLinks {
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
