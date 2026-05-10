package release

import "time"

// SearchCriteria はリリース検索条件
type SearchCriteria struct {
	Title           *string
	ReleaseType     *ReleaseType
	ArtistID        *string
	ArtistKind      *ArtistKind
	ReleaseDateFrom *time.Time
	ReleaseDateTo   *time.Time

	Sort  string
	Order string

	Offset int
	Limit  int
}
