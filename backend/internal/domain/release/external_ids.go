package release

import (
	"fmt"
	"regexp"
)

// ReleaseExternalIDKind はリリースの外部ID種別
type ReleaseExternalIDKind string

const (
	ReleaseExternalIDSpotifyAlbum    ReleaseExternalIDKind = "spotify_album_id"
	ReleaseExternalIDAppleMusicAlbum ReleaseExternalIDKind = "apple_music_album_id"
	ReleaseExternalIDUPC             ReleaseExternalIDKind = "upc"
	ReleaseExternalIDJANCode         ReleaseExternalIDKind = "jan_code"
)

var validReleaseExternalIDKinds = map[ReleaseExternalIDKind]struct{}{
	ReleaseExternalIDSpotifyAlbum:    {},
	ReleaseExternalIDAppleMusicAlbum: {},
	ReleaseExternalIDUPC:             {},
	ReleaseExternalIDJANCode:         {},
}

var validReleaseIDPattern = regexp.MustCompile(`^[a-zA-Z0-9._\-@]{1,256}$`)

// ReleaseExternalIDs はリリースの外部サービスIDマッピング（値オブジェクト）
type ReleaseExternalIDs struct {
	ids map[ReleaseExternalIDKind]string
}

func NewReleaseExternalIDs() *ReleaseExternalIDs {
	return &ReleaseExternalIDs{ids: make(map[ReleaseExternalIDKind]string)}
}

func ReconstructReleaseExternalIDs(ids map[ReleaseExternalIDKind]string) *ReleaseExternalIDs {
	if ids == nil {
		return NewReleaseExternalIDs()
	}
	copied := make(map[ReleaseExternalIDKind]string, len(ids))
	for k, v := range ids {
		copied[k] = v
	}
	return &ReleaseExternalIDs{ids: copied}
}

func (e *ReleaseExternalIDs) Get(kind ReleaseExternalIDKind) (string, bool) {
	v, ok := e.ids[kind]
	return v, ok
}

func (e *ReleaseExternalIDs) All() map[ReleaseExternalIDKind]string {
	result := make(map[ReleaseExternalIDKind]string, len(e.ids))
	for k, v := range e.ids {
		result[k] = v
	}
	return result
}

func (e *ReleaseExternalIDs) Set(kind ReleaseExternalIDKind, value string) error {
	if _, ok := validReleaseExternalIDKinds[kind]; !ok {
		return fmt.Errorf("無効な外部ID種別です: %s", kind)
	}
	if value == "" {
		delete(e.ids, kind)
		return nil
	}
	if !validReleaseIDPattern.MatchString(value) {
		return fmt.Errorf("外部IDの値が不正です（英数字・._-@のみ、256文字以内）: %s", kind)
	}
	e.ids[kind] = value
	return nil
}

func (e *ReleaseExternalIDs) Merge(other map[ReleaseExternalIDKind]string) error {
	for k, v := range other {
		if err := e.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (e *ReleaseExternalIDs) IsEmpty() bool {
	return len(e.ids) == 0
}

func ValidReleaseExternalIDKinds() []ReleaseExternalIDKind {
	kinds := make([]ReleaseExternalIDKind, 0, len(validReleaseExternalIDKinds))
	for k := range validReleaseExternalIDKinds {
		kinds = append(kinds, k)
	}
	return kinds
}
