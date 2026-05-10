package release

import "errors"

// GetReleaseQuery はリリース取得クエリ
type GetReleaseQuery struct {
	ID string
}

// ListReleasesQuery はリリース一覧取得クエリ
type ListReleasesQuery struct {
	Title           *string `form:"title"`
	ReleaseType     *string `form:"release_type"`
	ArtistID        *string `form:"artist_id"`
	ArtistKind      *string `form:"artist_kind"`
	ReleaseDateFrom *string `form:"release_date_from"` // YYYY-MM-DD
	ReleaseDateTo   *string `form:"release_date_to"`   // YYYY-MM-DD

	Sort  *string `form:"sort"`  // release_date, title, created_at
	Order *string `form:"order"` // asc, desc

	Page  *int `form:"page"`
	Limit *int `form:"limit"`
}

func (q *ListReleasesQuery) ApplyDefaults() {
	if q.Page == nil || *q.Page < 1 {
		v := 1
		q.Page = &v
	}
	if q.Limit == nil || *q.Limit < 1 {
		v := 20
		q.Limit = &v
	}
	if *q.Limit > 100 {
		v := 100
		q.Limit = &v
	}
	if q.Sort == nil {
		v := "release_date"
		q.Sort = &v
	}
	if q.Order == nil {
		v := "desc"
		q.Order = &v
	}
}

func (q *ListReleasesQuery) Validate() error {
	if q.ReleaseType != nil {
		allowed := []string{"single", "album", "ep", "mini_album", "digital_single", "compilation"}
		if !containsStr(allowed, *q.ReleaseType) {
			return errors.New("無効なリリース種別です")
		}
	}
	if q.ArtistKind != nil {
		allowed := []string{"idol", "group"}
		if !containsStr(allowed, *q.ArtistKind) {
			return errors.New("無効なアーティスト種別です")
		}
	}
	if q.Sort != nil {
		allowed := []string{"release_date", "title", "created_at"}
		if !containsStr(allowed, *q.Sort) {
			return errors.New("無効なソート項目です")
		}
	}
	if q.Order != nil {
		allowed := []string{"asc", "desc"}
		if !containsStr(allowed, *q.Order) {
			return errors.New("無効なソート順です")
		}
	}
	return nil
}

func containsStr(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ArtistRefDTO はアーティスト参照のデータ転送オブジェクト
type ArtistRefDTO struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
	Role string `json:"role,omitempty"`
}

// TrackDTO は収録曲のデータ転送オブジェクト
type TrackDTO struct {
	TrackNumber   int                   `json:"track_number"`
	Title         string                `json:"title"`
	DurationSec   *int                  `json:"duration_sec,omitempty"`
	ISRC          *string               `json:"isrc,omitempty"`
	CoverImageURL *string               `json:"cover_image_url,omitempty"`
	Participants  []TrackParticipantDTO `json:"participants,omitempty"`
}

// TrackParticipantDTO は楽曲単位のアイドル参加情報のデータ転送オブジェクト
type TrackParticipantDTO struct {
	IdolID   string  `json:"idol_id"`
	Status   string  `json:"status"`
	Position *string `json:"position,omitempty"`
}

// StreamingLinksDTO はストリーミングリンクのデータ転送オブジェクト
type StreamingLinksDTO struct {
	Spotify      *string `json:"spotify,omitempty"`
	AppleMusic   *string `json:"apple_music,omitempty"`
	YouTubeMusic *string `json:"youtube_music,omitempty"`
	YouTube      *string `json:"youtube,omitempty"`
	LineMusic    *string `json:"line_music,omitempty"`
	AmazonMusic  *string `json:"amazon_music,omitempty"`
	Official     *string `json:"official,omitempty"`
}

// ReleaseDTO はリリースのデータ転送オブジェクト
type ReleaseDTO struct {
	ID             string             `json:"id"`
	Title          string             `json:"title"`
	ReleaseType    string             `json:"release_type"`
	ReleaseDate    string             `json:"release_date"`
	Artists        []ArtistRefDTO     `json:"artists"`
	Tracks         []TrackDTO         `json:"tracks,omitempty"`
	StreamingLinks *StreamingLinksDTO `json:"streaming_links,omitempty"`
	CoverImageURL  *string            `json:"cover_image_url,omitempty"`
	Aliases        []string           `json:"aliases,omitempty"`
	TagIDs         []string           `json:"tag_ids,omitempty"`
	ExternalIDs    map[string]string  `json:"external_ids,omitempty"`
	CreatedAt      string             `json:"created_at"`
	UpdatedAt      string             `json:"updated_at"`
}

// SearchResult は検索結果のレスポンス構造
type SearchResult struct {
	Data  []*ReleaseDTO    `json:"data"`
	Meta  *PaginationMeta  `json:"meta"`
	Links *PaginationLinks `json:"links,omitempty"`
}

// PaginationMeta はページネーション情報
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// PaginationLinks はページネーションリンク
type PaginationLinks struct {
	First string  `json:"first"`
	Prev  *string `json:"prev"`
	Next  *string `json:"next"`
	Last  string  `json:"last"`
}
