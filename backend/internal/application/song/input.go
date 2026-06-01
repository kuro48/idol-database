package song

type CreateInput struct {
	Title         string
	TitleKana     *string
	DurationSec   *int
	ISRC          *string
	CoverImageURL *string
	Composers     []string
	Lyricists     []string
	Arrangers     []string
}

type UpdateInput struct {
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
