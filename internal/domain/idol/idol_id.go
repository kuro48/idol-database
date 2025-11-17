package idol

import (
	"errors"
)

// IdolID はアイドルの一意識別子
type IdolID struct {
	value string
}

func NewIdolID(value string) (IdolID, error) {
	if value == "" {
		return IdolID{}, errors.New("アイドルIDは空にできません")
	}
	return IdolID{value: value}, nil
}

func (id IdolID) Value() string {
	return id.value
}

func (id IdolID) Equals(other IdolID) bool {
	return id.value == other.value
}