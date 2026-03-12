package tag_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appTag "github.com/kuro48/idol-api/internal/application/tag"
	domaintag "github.com/kuro48/idol-api/internal/domain/tag"
	"github.com/kuro48/idol-api/internal/infrastructure/adapters"
	ucTag "github.com/kuro48/idol-api/internal/usecase/tag"
)

// inMemoryTagRepo はテスト用インメモリリポジトリ
type inMemoryTagRepo struct {
	data map[string]*domaintag.Tag
}

func newInMemoryTagRepo() *inMemoryTagRepo {
	return &inMemoryTagRepo{data: make(map[string]*domaintag.Tag)}
}

func (r *inMemoryTagRepo) Save(_ context.Context, t *domaintag.Tag) error {
	r.data[t.ID().String()] = t
	return nil
}

func (r *inMemoryTagRepo) Update(_ context.Context, t *domaintag.Tag) error {
	r.data[t.ID().String()] = t
	return nil
}

func (r *inMemoryTagRepo) Delete(_ context.Context, id domaintag.TagID) error {
	delete(r.data, id.String())
	return nil
}

func (r *inMemoryTagRepo) FindByID(_ context.Context, id domaintag.TagID) (*domaintag.Tag, error) {
	t, ok := r.data[id.String()]
	if !ok {
		return nil, fmt.Errorf("タグが見つかりません: %s", id.String())
	}
	return t, nil
}

func (r *inMemoryTagRepo) FindByName(_ context.Context, name string) (*domaintag.Tag, error) {
	for _, t := range r.data {
		if t.Name().String() == name {
			return t, nil
		}
	}
	return nil, nil
}

func (r *inMemoryTagRepo) FindByCategory(_ context.Context, cat domaintag.TagCategory) ([]*domaintag.Tag, error) {
	var result []*domaintag.Tag
	for _, t := range r.data {
		if t.Category() == cat {
			result = append(result, t)
		}
	}
	return result, nil
}

func (r *inMemoryTagRepo) Search(_ context.Context, criteria domaintag.SearchCriteria) ([]*domaintag.Tag, int64, error) {
	var result []*domaintag.Tag
	for _, t := range r.data {
		if criteria.Name != nil && t.Name().String() != *criteria.Name {
			continue
		}
		if criteria.Category != nil && t.Category() != *criteria.Category {
			continue
		}
		result = append(result, t)
	}
	return result, int64(len(result)), nil
}

func (r *inMemoryTagRepo) Exists(_ context.Context, id domaintag.TagID) (bool, error) {
	_, ok := r.data[id.String()]
	return ok, nil
}

func (r *inMemoryTagRepo) Restore(_ context.Context, id domaintag.TagID) error {
	if _, ok := r.data[id.String()]; !ok {
		return fmt.Errorf("削除済みタグが見つかりません: %s", id.String())
	}
	return nil
}

func newTagUsecase() ucTag.TagUseCase {
	repo := newInMemoryTagRepo()
	appSvc := appTag.NewApplicationService(repo)
	return ucTag.NewUsecase(adapters.NewTagAppAdapter(appSvc))
}

func TestCreateTag(t *testing.T) {
	uc := newTagUsecase()
	ctx := context.Background()

	dto, err := uc.CreateTag(ctx, ucTag.CreateTagCommand{
		Name:        "アイドル",
		Category:    "genre",
		Description: "アイドルジャンル",
	})

	require.NoError(t, err)
	assert.Equal(t, "アイドル", dto.Name)
	assert.Equal(t, "genre", dto.Category)
	assert.NotEmpty(t, dto.ID)
}

func TestCreateTag_DuplicateName(t *testing.T) {
	uc := newTagUsecase()
	ctx := context.Background()

	_, err := uc.CreateTag(ctx, ucTag.CreateTagCommand{Name: "重複", Category: "genre"})
	require.NoError(t, err)

	_, err = uc.CreateTag(ctx, ucTag.CreateTagCommand{Name: "重複", Category: "genre"})
	assert.Error(t, err)
}

func TestGetTag(t *testing.T) {
	uc := newTagUsecase()
	ctx := context.Background()

	created, err := uc.CreateTag(ctx, ucTag.CreateTagCommand{Name: "取得テスト", Category: "style"})
	require.NoError(t, err)

	got, err := uc.GetTag(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "取得テスト", got.Name)
}

func TestGetTag_NotFound(t *testing.T) {
	uc := newTagUsecase()
	ctx := context.Background()

	_, err := uc.GetTag(ctx, "000000000000000000000000")
	assert.Error(t, err)
}

func TestUpdateTag(t *testing.T) {
	uc := newTagUsecase()
	ctx := context.Background()

	created, err := uc.CreateTag(ctx, ucTag.CreateTagCommand{Name: "更新前", Category: "genre"})
	require.NoError(t, err)

	err = uc.UpdateTag(ctx, ucTag.UpdateTagCommand{
		ID:       created.ID,
		Name:     "更新後",
		Category: "style",
	})
	require.NoError(t, err)

	got, err := uc.GetTag(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "更新後", got.Name)
	assert.Equal(t, "style", got.Category)
}

func TestDeleteTag(t *testing.T) {
	uc := newTagUsecase()
	ctx := context.Background()

	created, err := uc.CreateTag(ctx, ucTag.CreateTagCommand{Name: "削除対象", Category: "other"})
	require.NoError(t, err)

	err = uc.DeleteTag(ctx, created.ID)
	require.NoError(t, err)

	_, err = uc.GetTag(ctx, created.ID)
	assert.Error(t, err)
}

func TestSearchTags(t *testing.T) {
	uc := newTagUsecase()
	ctx := context.Background()

	_, err := uc.CreateTag(ctx, ucTag.CreateTagCommand{Name: "タグA", Category: "genre"})
	require.NoError(t, err)
	_, err = uc.CreateTag(ctx, ucTag.CreateTagCommand{Name: "タグB", Category: "region"})
	require.NoError(t, err)

	result, err := uc.SearchTags(ctx, ucTag.SearchQuery{Page: 1, Limit: 20}, "/api/v1/tags")
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Meta.Total)
}
