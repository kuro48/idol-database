package release

import (
	"errors"
	"fmt"
	"time"
)

// Release は音楽リリース集約のルートエンティティ
type Release struct {
	id             ReleaseID
	title          ReleaseTitle
	releaseType    ReleaseType
	releaseDate    ReleaseDate
	artists        []ArtistRef
	tracks         []Track
	streamingLinks *StreamingLinks
	externalIDs    *ReleaseExternalIDs
	coverImageURL  *string
	aliases        []string
	tagIDs         []string
	createdAt      time.Time
	updatedAt      time.Time
}

// NewRelease は新しいリリースを作成する
func NewRelease(
	title ReleaseTitle,
	releaseType ReleaseType,
	releaseDate ReleaseDate,
	artists []ArtistRef,
) (*Release, error) {
	if len(artists) == 0 {
		return nil, errors.New("リリースにはアーティストが1名以上必要です")
	}
	now := time.Now()
	return &Release{
		title:       title,
		releaseType: releaseType,
		releaseDate: releaseDate,
		artists:     artists,
		tracks:      []Track{},
		tagIDs:      []string{},
		aliases:     []string{},
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// Reconstruct はデータストアからリリースを再構築する（永続化層用）
func Reconstruct(
	id ReleaseID,
	title ReleaseTitle,
	releaseType ReleaseType,
	releaseDate ReleaseDate,
	artists []ArtistRef,
	tracks []Track,
	streamingLinks *StreamingLinks,
	externalIDs *ReleaseExternalIDs,
	coverImageURL *string,
	aliases []string,
	tagIDs []string,
	createdAt time.Time,
	updatedAt time.Time,
) *Release {
	return &Release{
		id:             id,
		title:          title,
		releaseType:    releaseType,
		releaseDate:    releaseDate,
		artists:        artists,
		tracks:         tracks,
		streamingLinks: streamingLinks,
		externalIDs:    externalIDs,
		coverImageURL:  coverImageURL,
		aliases:        aliases,
		tagIDs:         tagIDs,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

func (r *Release) ID() ReleaseID           { return r.id }
func (r *Release) Title() ReleaseTitle     { return r.title }
func (r *Release) ReleaseType() ReleaseType { return r.releaseType }
func (r *Release) ReleaseDate() ReleaseDate { return r.releaseDate }
func (r *Release) Artists() []ArtistRef    { return r.artists }
func (r *Release) Tracks() []Track         { return r.tracks }
func (r *Release) CoverImageURL() *string  { return r.coverImageURL }
func (r *Release) CreatedAt() time.Time    { return r.createdAt }
func (r *Release) UpdatedAt() time.Time    { return r.updatedAt }

func (r *Release) StreamingLinks() *StreamingLinks {
	if r.streamingLinks == nil {
		return NewStreamingLinks()
	}
	return r.streamingLinks
}

func (r *Release) ExternalIDs() *ReleaseExternalIDs {
	if r.externalIDs == nil {
		return NewReleaseExternalIDs()
	}
	return r.externalIDs
}

func (r *Release) Aliases() []string {
	if r.aliases == nil {
		return []string{}
	}
	return r.aliases
}

func (r *Release) TagIDs() []string {
	if r.tagIDs == nil {
		return []string{}
	}
	return r.tagIDs
}

func (r *Release) SetID(id ReleaseID) {
	r.id = id
}

func (r *Release) ChangeTitle(title ReleaseTitle) error {
	if title.Value() == "" {
		return errors.New("タイトルは空にできません")
	}
	r.title = title
	r.updatedAt = time.Now()
	return nil
}

func (r *Release) UpdateType(t ReleaseType) {
	r.releaseType = t
	r.updatedAt = time.Now()
}

func (r *Release) UpdateReleaseDate(d ReleaseDate) {
	r.releaseDate = d
	r.updatedAt = time.Now()
}

// SetArtists はアーティスト参照リストを設定する（1件以上必須）
func (r *Release) SetArtists(artists []ArtistRef) error {
	if len(artists) == 0 {
		return errors.New("リリースにはアーティストが1名以上必要です")
	}
	r.artists = artists
	r.updatedAt = time.Now()
	return nil
}

// SetTracks は収録曲リストを設定する（trackNumber の一意性を検証）
func (r *Release) SetTracks(tracks []Track) error {
	seen := make(map[int]struct{}, len(tracks))
	for _, t := range tracks {
		if _, exists := seen[t.TrackNumber()]; exists {
			return fmt.Errorf("トラック番号が重複しています: %d", t.TrackNumber())
		}
		seen[t.TrackNumber()] = struct{}{}
	}
	r.tracks = tracks
	r.updatedAt = time.Now()
	return nil
}

func (r *Release) UpdateStreamingLinks(links *StreamingLinks) {
	r.streamingLinks = links
	r.updatedAt = time.Now()
}

func (r *Release) UpdateExternalIDs(ids *ReleaseExternalIDs) {
	r.externalIDs = ids
	r.updatedAt = time.Now()
}

func (r *Release) SetCoverImageURL(url *string) {
	r.coverImageURL = url
	r.updatedAt = time.Now()
}

func (r *Release) SetAliases(aliases []string) {
	r.aliases = aliases
	r.updatedAt = time.Now()
}

func (r *Release) SetTags(tagIDs []string) {
	r.tagIDs = tagIDs
	r.updatedAt = time.Now()
}

// Validate はリリースの状態が有効かを検証する
func (r *Release) Validate() error {
	if r.title.Value() == "" {
		return NewDomainError("タイトルは必須です")
	}
	if !r.releaseType.IsValid() {
		return NewDomainError("無効なリリース種別です")
	}
	if len(r.artists) == 0 {
		return NewDomainError("アーティストは1名以上必要です")
	}
	return nil
}
