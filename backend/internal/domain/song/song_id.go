package song

import "errors"

// SongID は曲のIDを表す値オブジェクト
type SongID struct {
	value string
}

func NewSongID(value string) (SongID, error) {
	if value == "" {
		return SongID{}, errors.New("曲IDは空にできません")
	}
	return SongID{value: value}, nil
}

func (id SongID) Value() string {
	return id.value
}

func (id SongID) Equals(other SongID) bool {
	return id.value == other.value
}
