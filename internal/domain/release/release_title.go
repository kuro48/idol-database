package release

import "errors"

// ReleaseTitle はリリースのタイトル
type ReleaseTitle struct {
	value string
}

func NewReleaseTitle(value string) (ReleaseTitle, error) {
	if value == "" {
		return ReleaseTitle{}, errors.New("リリースタイトルは空にできません")
	}
	if len([]rune(value)) > 200 {
		return ReleaseTitle{}, errors.New("リリースタイトルは200文字以内である必要があります")
	}
	return ReleaseTitle{value: value}, nil
}

func (t ReleaseTitle) Value() string {
	return t.value
}
