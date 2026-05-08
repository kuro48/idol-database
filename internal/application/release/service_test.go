package release

import (
	"context"
	"fmt"
	"testing"
	"time"

	domainRelease "github.com/kuro48/idol-api/internal/domain/release"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type inMemoryReleaseRepo struct {
	data map[string]*domainRelease.Release
}

func newInMemoryReleaseRepo() *inMemoryReleaseRepo {
	return &inMemoryReleaseRepo{data: make(map[string]*domainRelease.Release)}
}

func (r *inMemoryReleaseRepo) Save(_ context.Context, rel *domainRelease.Release) error {
	if rel.ID().Value() == "" {
		id, err := domainRelease.NewReleaseID(fmt.Sprintf("release-%d", len(r.data)+1))
		if err != nil {
			return err
		}
		rel.SetID(id)
	}
	r.data[rel.ID().Value()] = rel
	return nil
}

func (r *inMemoryReleaseRepo) FindByID(_ context.Context, id domainRelease.ReleaseID) (*domainRelease.Release, error) {
	rel, ok := r.data[id.Value()]
	if !ok {
		return nil, fmt.Errorf("リリースが見つかりません: %s", id.Value())
	}
	return rel, nil
}

func (r *inMemoryReleaseRepo) Update(_ context.Context, rel *domainRelease.Release) error {
	r.data[rel.ID().Value()] = rel
	return nil
}

func (r *inMemoryReleaseRepo) Delete(_ context.Context, id domainRelease.ReleaseID) error {
	delete(r.data, id.Value())
	return nil
}

func (r *inMemoryReleaseRepo) Restore(_ context.Context, id domainRelease.ReleaseID) error {
	if _, ok := r.data[id.Value()]; !ok {
		return fmt.Errorf("削除済みリリースが見つかりません: %s", id.Value())
	}
	return nil
}

func (r *inMemoryReleaseRepo) Search(_ context.Context, criteria domainRelease.SearchCriteria) ([]*domainRelease.Release, error) {
	result := make([]*domainRelease.Release, 0, len(r.data))
	for _, rel := range r.data {
		if criteria.Title != nil && rel.Title().Value() != *criteria.Title {
			continue
		}
		if criteria.ReleaseType != nil && rel.ReleaseType() != *criteria.ReleaseType {
			continue
		}
		result = append(result, rel)
	}
	return result, nil
}

func (r *inMemoryReleaseRepo) Count(ctx context.Context, criteria domainRelease.SearchCriteria) (int64, error) {
	result, err := r.Search(ctx, criteria)
	if err != nil {
		return 0, err
	}
	return int64(len(result)), nil
}

func (r *inMemoryReleaseRepo) FindByExternalID(_ context.Context, _ domainRelease.ReleaseExternalIDKind, _ string) (*domainRelease.Release, error) {
	return nil, nil
}

func (r *inMemoryReleaseRepo) FindByArtistID(_ context.Context, artistID string) ([]*domainRelease.Release, error) {
	result := []*domainRelease.Release{}
	for _, rel := range r.data {
		for _, artist := range rel.Artists() {
			if artist.ID() == artistID {
				result = append(result, rel)
				break
			}
		}
	}
	return result, nil
}

func TestCreateReleaseRejectsDuplicateTrackNumbers(t *testing.T) {
	svc := NewApplicationService(newInMemoryReleaseRepo(), nil)

	_, err := svc.CreateRelease(context.Background(), CreateInput{
		Title:       "重複トラック",
		ReleaseType: "single",
		ReleaseDate: "2026-05-08",
		Artists: []ArtistRefInput{
			{Kind: "group", ID: "group-1", Role: "main"},
		},
		Tracks: []TrackInput{
			{TrackNumber: 1, Title: "表題曲"},
			{TrackNumber: 1, Title: "別曲"},
		},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "トラック番号が重複しています")
}

func TestUpdateReleaseUpdatesStreamingLinks(t *testing.T) {
	repo := newInMemoryReleaseRepo()
	svc := NewApplicationService(repo, nil)

	created, err := svc.CreateRelease(context.Background(), CreateInput{
		Title:       "配信リンク更新",
		ReleaseType: "single",
		ReleaseDate: "2026-05-08",
		Artists: []ArtistRefInput{
			{Kind: "idol", ID: "idol-1", Role: "main"},
		},
		Tracks: []TrackInput{
			{TrackNumber: 1, Title: "表題曲"},
		},
	})
	require.NoError(t, err)

	spotify := "https://open.spotify.com/album/example"
	appleMusic := "https://music.apple.com/jp/album/example"
	err = svc.UpdateRelease(context.Background(), UpdateInput{
		ID: created.ID().Value(),
		StreamingLinks: &StreamingLinksInput{
			Spotify:    &spotify,
			AppleMusic: &appleMusic,
		},
	})
	require.NoError(t, err)

	got, err := svc.GetRelease(context.Background(), created.ID().Value())
	require.NoError(t, err)
	require.NotNil(t, got.StreamingLinks())
	assert.Equal(t, spotify, *got.StreamingLinks().Spotify())
	assert.Equal(t, appleMusic, *got.StreamingLinks().AppleMusic())
	assert.WithinDuration(t, time.Now(), got.UpdatedAt(), time.Second)
}
