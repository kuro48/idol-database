package song

import "errors"

type GetSongQuery struct {
	ID string
}

type ListSongQuery struct {
	Title *string `form:"title"`
	ISRC  *string `form:"isrc"`
	Sort  *string `form:"sort"`
	Order *string `form:"order"`
	Page  *int    `form:"page"`
	Limit *int    `form:"limit"`
}

func (q *ListSongQuery) Normalize() {
	if q.Page == nil || *q.Page < 1 {
		p := 1
		q.Page = &p
	}
	if q.Limit == nil || *q.Limit < 1 {
		l := 20
		q.Limit = &l
	}
	if *q.Limit > 100 {
		l := 100
		q.Limit = &l
	}
	if q.Sort == nil {
		s := "created_at"
		q.Sort = &s
	}
	if q.Order == nil {
		o := "desc"
		q.Order = &o
	}
}

func (q *ListSongQuery) Validate() error {
	if q.Sort != nil {
		allowed := []string{"title", "created_at"}
		if !songContainsStr(allowed, *q.Sort) {
			return errors.New("無効なソート項目です")
		}
	}
	if q.Order != nil {
		if *q.Order != "asc" && *q.Order != "desc" {
			return errors.New("無効なソート順です")
		}
	}
	return nil
}

func songContainsStr(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

type SongDTO struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	TitleKana     *string  `json:"title_kana,omitempty"`
	DurationSec   *int     `json:"duration_sec,omitempty"`
	ISRC          *string  `json:"isrc,omitempty"`
	CoverImageURL *string  `json:"cover_image_url,omitempty"`
	Composers     []string `json:"composers"`
	Lyricists     []string `json:"lyricists"`
	Arrangers     []string `json:"arrangers"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

type SongSearchResult struct {
	Data []*SongDTO      `json:"data"`
	Meta *SongPagination `json:"meta"`
}

type SongPagination struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}
