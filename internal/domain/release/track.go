package release

import (
	"errors"
	"fmt"
)

// Track はリリースに収録された楽曲（値オブジェクト）
type Track struct {
	trackNumber   int
	title         string
	durationSec   *int    // 再生時間（秒）、省略可
	isrc          *string // International Standard Recording Code、省略可
	coverImageURL *string // 楽曲個別のジャケット画像URL、省略可
}

func NewTrack(trackNumber int, title string, durationSec *int, isrc *string, coverImageURL *string) (Track, error) {
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
	return Track{
		trackNumber:   trackNumber,
		title:         title,
		durationSec:   durationSec,
		isrc:          isrc,
		coverImageURL: coverImageURL,
	}, nil
}

func (t Track) TrackNumber() int      { return t.trackNumber }
func (t Track) Title() string         { return t.title }
func (t Track) DurationSec() *int     { return t.durationSec }
func (t Track) ISRC() *string         { return t.isrc }
func (t Track) CoverImageURL() *string { return t.coverImageURL }
