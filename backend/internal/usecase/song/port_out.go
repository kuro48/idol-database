package song

import (
	"context"

	domain "github.com/kuro48/idol-api/internal/domain/song"
)

// SongAppPort は Usecase が application サービスに要求する契約
type SongAppPort interface {
	CreateSong(ctx context.Context, input SongCreateInput) (*domain.Song, error)
	GetSong(ctx context.Context, id string) (*domain.Song, error)
	SearchSongs(ctx context.Context, criteria domain.SearchCriteria) ([]*domain.Song, error)
	CountSongs(ctx context.Context, criteria domain.SearchCriteria) (int64, error)
	UpdateSong(ctx context.Context, input SongUpdateInput) error
	DeleteSong(ctx context.Context, id string) error
}

type SongCreateInput struct {
	Title         string
	TitleKana     *string
	DurationSec   *int
	ISRC          *string
	CoverImageURL *string
	Composers     []string
	Lyricists     []string
	Arrangers     []string
}

type SongUpdateInput struct {
	ID            string
	Title         string
	TitleKana     *string
	DurationSec   *int
	ISRC          *string
	CoverImageURL *string
	Composers     []string
	Lyricists     []string
	Arrangers     []string
}
