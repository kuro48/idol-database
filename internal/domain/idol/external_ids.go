package idol

import (
	"errors"
	"fmt"
	"regexp"
)

// ExternalIDKind は外部IDの種別
type ExternalIDKind string

const (
	ExternalIDKindTwitter       ExternalIDKind = "twitter"
	ExternalIDKindInstagram     ExternalIDKind = "instagram"
	ExternalIDKindTikTok        ExternalIDKind = "tiktok"
	ExternalIDKindYouTube       ExternalIDKind = "youtube_channel"
	ExternalIDKindSpotify       ExternalIDKind = "spotify_artist"
	ExternalIDKindAppleMusic    ExternalIDKind = "apple_music_artist"
	ExternalIDKindAmebaAmeba    ExternalIDKind = "ameba"
	ExternalIDKindNote          ExternalIDKind = "note"
	ExternalIDKindWikipediaJa   ExternalIDKind = "wikipedia_ja"
	ExternalIDKindWikipediaEn   ExternalIDKind = "wikipedia_en"
)

// validKinds は受け入れ可能な外部ID種別の集合
var validKinds = map[ExternalIDKind]struct{}{
	ExternalIDKindTwitter:     {},
	ExternalIDKindInstagram:   {},
	ExternalIDKindTikTok:      {},
	ExternalIDKindYouTube:     {},
	ExternalIDKindSpotify:     {},
	ExternalIDKindAppleMusic:  {},
	ExternalIDKindAmebaAmeba:  {},
	ExternalIDKindNote:        {},
	ExternalIDKindWikipediaJa: {},
	ExternalIDKindWikipediaEn: {},
}

// validIDPattern は外部IDの値として許容するパターン（英数字・アンダースコア・ハイフン・ドット）
var validIDPattern = regexp.MustCompile(`^[a-zA-Z0-9._\-@]{1,256}$`)

// ExternalIDs はアイドルの外部サービスIDマッピングを表す値オブジェクト
type ExternalIDs struct {
	ids map[ExternalIDKind]string
}

// NewExternalIDs は空の ExternalIDs を返す
func NewExternalIDs() *ExternalIDs {
	return &ExternalIDs{ids: make(map[ExternalIDKind]string)}
}

// ReconstructExternalIDs は永続化層からの再構築用
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

// Get は指定種別の外部IDを返す
func (e *ExternalIDs) Get(kind ExternalIDKind) (string, bool) {
	v, ok := e.ids[kind]
	return v, ok
}

// All は全外部IDのコピーを返す
func (e *ExternalIDs) All() map[ExternalIDKind]string {
	result := make(map[ExternalIDKind]string, len(e.ids))
	for k, v := range e.ids {
		result[k] = v
	}
	return result
}

// Set は外部IDを設定する（空文字列で削除）
func (e *ExternalIDs) Set(kind ExternalIDKind, value string) error {
	if _, ok := validKinds[kind]; !ok {
		return fmt.Errorf("無効な外部ID種別です: %s", kind)
	}
	if value == "" {
		delete(e.ids, kind)
		return nil
	}
	if !validIDPattern.MatchString(value) {
		return errors.New("外部IDの値が不正です（英数字・._-@のみ、256文字以内）")
	}
	e.ids[kind] = value
	return nil
}

// Merge は別の ExternalIDs をマージする（上書き）
func (e *ExternalIDs) Merge(other map[ExternalIDKind]string) error {
	for k, v := range other {
		if err := e.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

// IsEmpty は空かどうかを返す
func (e *ExternalIDs) IsEmpty() bool {
	return len(e.ids) == 0
}

// ValidKinds は有効な外部ID種別のリストを返す
func ValidExternalIDKinds() []ExternalIDKind {
	kinds := make([]ExternalIDKind, 0, len(validKinds))
	for k := range validKinds {
		kinds = append(kinds, k)
	}
	return kinds
}
