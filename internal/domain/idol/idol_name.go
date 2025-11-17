package idol

import (
	"errors"
)

// IdolName はアイドルの名前
type IdolName struct {
	value string
}

func NewIdolName(value string) (IdolName, error) {
	if value == "" {
		return IdolName{}, errors.New("アイドル名は空にできません")
	}
	if len(value) > 100 {
		return IdolName{}, errors.New("アイドル名は100文字以内である必要があります")
	}
	return IdolName{value: value}, nil
}

func (n IdolName) Value() string {
	return n.value
}