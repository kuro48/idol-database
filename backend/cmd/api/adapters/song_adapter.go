package adapters

import (
	"context"

	appSong "github.com/kuro48/idol-api/internal/application/song"
	domainSong "github.com/kuro48/idol-api/internal/domain/song"
	ucSong "github.com/kuro48/idol-api/internal/usecase/song"
)

type SongAppAdapter struct {
	svc *appSong.ApplicationService
}

func NewSongAppAdapter(svc *appSong.ApplicationService) ucSong.SongAppPort {
	return &SongAppAdapter{svc: svc}
}

func (a *SongAppAdapter) CreateSong(ctx context.Context, input ucSong.SongCreateInput) (*domainSong.Song, error) {
	return a.svc.CreateSong(ctx, appSong.CreateInput{
		Title:         input.Title,
		TitleKana:     input.TitleKana,
		DurationSec:   input.DurationSec,
		ISRC:          input.ISRC,
		CoverImageURL: input.CoverImageURL,
		Composers:     input.Composers,
		Lyricists:     input.Lyricists,
		Arrangers:     input.Arrangers,
	})
}

func (a *SongAppAdapter) GetSong(ctx context.Context, id string) (*domainSong.Song, error) {
	return a.svc.GetSong(ctx, id)
}

func (a *SongAppAdapter) SearchSongs(ctx context.Context, criteria domainSong.SearchCriteria) ([]*domainSong.Song, error) {
	return a.svc.SearchSongs(ctx, criteria)
}

func (a *SongAppAdapter) CountSongs(ctx context.Context, criteria domainSong.SearchCriteria) (int64, error) {
	return a.svc.CountSongs(ctx, criteria)
}

func (a *SongAppAdapter) UpdateSong(ctx context.Context, input ucSong.SongUpdateInput) error {
	return a.svc.UpdateSong(ctx, appSong.UpdateInput{
		ID:            input.ID,
		Title:         input.Title,
		TitleKana:     input.TitleKana,
		DurationSec:   input.DurationSec,
		ISRC:          input.ISRC,
		CoverImageURL: input.CoverImageURL,
		Composers:     input.Composers,
		Lyricists:     input.Lyricists,
		Arrangers:     input.Arrangers,
	})
}

func (a *SongAppAdapter) DeleteSong(ctx context.Context, id string) error {
	return a.svc.DeleteSong(ctx, id)
}
