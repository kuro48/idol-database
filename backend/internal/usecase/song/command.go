package song

type CreateSongCommand struct {
	Title         string
	TitleKana     *string
	DurationSec   *int
	ISRC          *string
	CoverImageURL *string
	Composers     []string
	Lyricists     []string
	Arrangers     []string
}

type UpdateSongCommand struct {
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

type DeleteSongCommand struct {
	ID string
}
