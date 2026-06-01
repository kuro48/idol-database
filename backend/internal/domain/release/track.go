package release

import (
	"errors"
	"fmt"
)

// Track はリリースに収録された楽曲（値オブジェクト）
type Track struct {
	trackNumber   int
	title         string
	titleKana     *string
	durationSec   *int    // 再生時間（秒）、省略可
	isrc          *string // International Standard Recording Code、省略可
	coverImageURL *string // 楽曲個別のジャケット画像URL、省略可
	composers     []string
	lyricists     []string
	arrangers     []string
	participants  []TrackParticipant
}

func NewTrack(trackNumber int, title string, titleKana *string, durationSec *int, isrc *string, coverImageURL *string, composers, lyricists, arrangers []string, participants []TrackParticipant) (Track, error) {
	if trackNumber < 1 {
		return Track{}, fmt.Errorf("トラック番号は1以上である必要があります: %d", trackNumber)
	}
	if title == "" {
		return Track{}, errors.New("楽曲タイトルは空にできません")
	}
	if len([]rune(title)) > 200 {
		return Track{}, errors.New("楽曲タイトルは200文字以内である必要があります")
	}
	if durationSec != nil && *durationSec < 0 {
		return Track{}, errors.New("再生時間は0以上である必要があります")
	}
	seen := make(map[string]struct{}, len(participants))
	for _, p := range participants {
		if _, exists := seen[p.IdolID()]; exists {
			return Track{}, fmt.Errorf("楽曲参加アイドルIDが重複しています: %s", p.IdolID())
		}
		seen[p.IdolID()] = struct{}{}
	}
	return Track{
		trackNumber:   trackNumber,
		title:         title,
		titleKana:     titleKana,
		durationSec:   durationSec,
		isrc:          isrc,
		coverImageURL: coverImageURL,
		composers:     composers,
		lyricists:     lyricists,
		arrangers:     arrangers,
		participants:  participants,
	}, nil
}

func (t Track) TrackNumber() int       { return t.trackNumber }
func (t Track) Title() string          { return t.title }
func (t Track) TitleKana() *string     { return t.titleKana }
func (t Track) DurationSec() *int      { return t.durationSec }
func (t Track) ISRC() *string          { return t.isrc }
func (t Track) CoverImageURL() *string { return t.coverImageURL }
func (t Track) Composers() []string {
	if t.composers == nil {
		return []string{}
	}
	return t.composers
}
func (t Track) Lyricists() []string {
	if t.lyricists == nil {
		return []string{}
	}
	return t.lyricists
}
func (t Track) Arrangers() []string {
	if t.arrangers == nil {
		return []string{}
	}
	return t.arrangers
}
func (t Track) Participants() []TrackParticipant {
	if t.participants == nil {
		return []TrackParticipant{}
	}
	return t.participants
}
