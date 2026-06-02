package release

// ArtistRefInput はアーティスト参照の入力
type ArtistRefInput struct {
	Kind string // "idol" or "group"
	ID   string
	Role string
}

// TrackInput は収録曲の入力
type TrackInput struct {
	TrackNumber   int
	Title         string
	TitleKana     *string
	DurationSec   *int
	ISRC          *string
	CoverImageURL *string
	Composers     []string
	Lyricists     []string
	Arrangers     []string
	Participants  []TrackParticipantInput
}

// TrackParticipantInput は楽曲単位のアイドル参加情報入力
type TrackParticipantInput struct {
	IdolID   string
	Status   string // "participating" or "not_participating"
	Position *string
}

// StreamingLinksInput はストリーミングリンクの入力
type StreamingLinksInput struct {
	Spotify      *string
	AppleMusic   *string
	YouTubeMusic *string
	YouTube      *string
	LineMusic    *string
	AmazonMusic  *string
	Official     *string
}

// CreateInput はリリース作成の入力
type CreateInput struct {
	Title          string
	ReleaseType    string
	ReleaseDate    string // "YYYY-MM-DD"
	Artists        []ArtistRefInput
	Tracks         []TrackInput
	StreamingLinks *StreamingLinksInput
	CoverImageURL  *string
	Aliases        []string
	TagIDs         []string
}

// UpdateInput はリリース更新の入力
type UpdateInput struct {
	ID             string
	Title          *string
	ReleaseType    *string
	ReleaseDate    *string
	Artists        []ArtistRefInput
	Tracks         []TrackInput
	StreamingLinks *StreamingLinksInput
	CoverImageURL  *string
	Aliases        []string
	TagIDs         []string
}

// UpdateStreamingLinksInput はストリーミングリンク更新の入力
type UpdateStreamingLinksInput struct {
	ID    string
	Links StreamingLinksInput
}

// UpdateExternalIDsInput は外部ID更新の入力
type UpdateExternalIDsInput struct {
	ID          string
	ExternalIDs map[string]string // キーは ReleaseExternalIDKind の文字列値
}
