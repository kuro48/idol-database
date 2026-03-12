package adapters

import (
	"context"

	appTag "github.com/kuro48/idol-api/internal/application/tag"
	tagDomain "github.com/kuro48/idol-api/internal/domain/tag"
	ucTag "github.com/kuro48/idol-api/internal/usecase/tag"
)

// TagAppAdapter は appTag.ApplicationService を ucTag.TagAppPort に適合させる
type TagAppAdapter struct {
	svc *appTag.ApplicationService
}

// NewTagAppAdapter は TagAppAdapter を生成する
func NewTagAppAdapter(svc *appTag.ApplicationService) ucTag.TagAppPort {
	return &TagAppAdapter{svc: svc}
}

func (a *TagAppAdapter) CreateTag(ctx context.Context, input ucTag.TagCreateInput) (*tagDomain.Tag, error) {
	return a.svc.CreateTag(ctx, appTag.CreateInput{
		Name:        input.Name,
		Category:    input.Category,
		Description: input.Description,
	})
}

func (a *TagAppAdapter) UpdateTag(ctx context.Context, input ucTag.TagUpdateInput) error {
	return a.svc.UpdateTag(ctx, appTag.UpdateInput{
		ID:          input.ID,
		Name:        input.Name,
		Category:    input.Category,
		Description: input.Description,
	})
}

func (a *TagAppAdapter) DeleteTag(ctx context.Context, id string) error {
	return a.svc.DeleteTag(ctx, id)
}

func (a *TagAppAdapter) GetTag(ctx context.Context, id string) (*tagDomain.Tag, error) {
	return a.svc.GetTag(ctx, id)
}

func (a *TagAppAdapter) SearchTags(ctx context.Context, criteria tagDomain.SearchCriteria) ([]*tagDomain.Tag, int64, error) {
	return a.svc.SearchTags(ctx, criteria)
}
