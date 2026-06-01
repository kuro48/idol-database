package group

import (
	"errors"
	"fmt"
	"regexp"
)

// ExternalIDKind はグループの外部ID種別
type ExternalIDKind string

const (
	ExternalIDKindMusicBrainz ExternalIDKind = "musicbrainz_artist"
	ExternalIDKindSpotify     ExternalIDKind = "spotify_artist"
	ExternalIDKindWikidata    ExternalIDKind = "wikidata"
	ExternalIDKindWikipediaJa ExternalIDKind = "wikipedia_ja"
	ExternalIDKindWikipediaEn ExternalIDKind = "wikipedia_en"
)

var validGroupKinds = map[ExternalIDKind]struct{}{
	ExternalIDKindMusicBrainz: {},
	ExternalIDKindSpotify:     {},
	ExternalIDKindWikidata:    {},
	ExternalIDKindWikipediaJa: {},
	ExternalIDKindWikipediaEn: {},
}

var validGroupIDPattern = regexp.MustCompile(`^[a-zA-Z0-9._\-@/]{1,256}$`)

// ExternalIDs はグループの外部サービスIDマッピングを表す値オブジェクト
type ExternalIDs struct {
	ids map[ExternalIDKind]string
}

func NewExternalIDs() *ExternalIDs {
	return &ExternalIDs{ids: make(map[ExternalIDKind]string)}
}

func ReconstructExternalIDs(ids map[ExternalIDKind]string) *ExternalIDs {
	if ids == nil {
		return NewExternalIDs()
	}
	copied := make(map[ExternalIDKind]string, len(ids))
	for k, v := range ids {
		copied[k] = v
	}
	return &ExternalIDs{ids: copied}
}

func (e *ExternalIDs) Get(kind ExternalIDKind) (string, bool) {
	v, ok := e.ids[kind]
	return v, ok
}

func (e *ExternalIDs) All() map[ExternalIDKind]string {
	result := make(map[ExternalIDKind]string, len(e.ids))
	for k, v := range e.ids {
		result[k] = v
	}
	return result
}

func (e *ExternalIDs) Set(kind ExternalIDKind, value string) error {
	if _, ok := validGroupKinds[kind]; !ok {
		return fmt.Errorf("無効な外部ID種別です: %s", kind)
	}
	if value == "" {
		delete(e.ids, kind)
		return nil
	}
	if !validGroupIDPattern.MatchString(value) {
		return errors.New("外部IDの値が不正です（英数字・._-@/のみ、256文字以内）")
	}
	e.ids[kind] = value
	return nil
}

func (e *ExternalIDs) Merge(other map[ExternalIDKind]string) error {
	for k, v := range other {
		if err := e.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (e *ExternalIDs) IsEmpty() bool {
	return len(e.ids) == 0
}

func ValidExternalIDKinds() []ExternalIDKind {
	kinds := make([]ExternalIDKind, 0, len(validGroupKinds))
	for k := range validGroupKinds {
		kinds = append(kinds, k)
	}
	return kinds
}
