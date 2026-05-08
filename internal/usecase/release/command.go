package release

// ArtistRefCommand はアーティスト参照のコマンド入力
type ArtistRefCommand struct {
	Kind string // "idol" or "group"
	ID   string
	Role string
}

// TrackCommand は収録曲のコマンド入力
type TrackCommand struct {
	TrackNumber   int
	Title         string
	DurationSec   *int
	ISRC          *string
	CoverImageURL *string
}

// StreamingLinksCommand はストリーミングリンクのコマンド入力
type StreamingLinksCommand struct {
	Spotify      *string `json:"spotify"`
	AppleMusic   *string `json:"apple_music"`
	YouTubeMusic *string `json:"youtube_music"`
	YouTube      *string `json:"youtube"`
	LineMusic    *string `json:"line_music"`
	AmazonMusic  *string `json:"amazon_music"`
	Official     *string `json:"official"`
}

// CreateReleaseCommand はリリース作成コマンド
type CreateReleaseCommand struct {
	Title          string
	ReleaseType    string
	ReleaseDate    string // "YYYY-MM-DD"
	Artists        []ArtistRefCommand
	Tracks         []TrackCommand
	StreamingLinks *StreamingLinksCommand
	CoverImageURL  *string
	Aliases        []string
	TagIDs         []string
}

// UpdateReleaseCommand はリリース更新コマンド
type UpdateReleaseCommand struct {
	ID             string
	Title          *string
	ReleaseType    *string
	ReleaseDate    *string
	Artists        []ArtistRefCommand
	Tracks         []TrackCommand
	StreamingLinks *StreamingLinksCommand
	CoverImageURL  *string
	Aliases        []string
	TagIDs         []string
}

// DeleteReleaseCommand はリリース削除コマンド
type DeleteReleaseCommand struct {
	ID string
}

// UpdateStreamingLinksCommand はストリーミングリンク更新コマンド
type UpdateStreamingLinksCommand struct {
	ID    string
	Links StreamingLinksCommand
}

// UpdateExternalIDsCommand は外部ID更新コマンド
type UpdateExternalIDsCommand struct {
	ID          string
	ExternalIDs map[string]string
}
