package release

import (
	"errors"
	"time"
)

// ReleaseDate はリリース日（発売日・配信日）
type ReleaseDate struct {
	value time.Time
}

// NewReleaseDateFromString は "YYYY-MM-DD" 形式の文字列からReleaseDateを作成する
// 未来日付は予告対応のため許可する
func NewReleaseDateFromString(s string) (ReleaseDate, error) {
	if s == "" {
		return ReleaseDate{}, errors.New("リリース日は空にできません")
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return ReleaseDate{}, errors.New("リリース日はYYYY-MM-DD形式で指定してください")
	}
	return ReleaseDate{value: t.UTC()}, nil
}

// NewReleaseDate は time.Time からReleaseDateを作成する（永続化層用）
func NewReleaseDate(t time.Time) ReleaseDate {
	return ReleaseDate{value: t.UTC()}
}

func (d ReleaseDate) Value() time.Time {
	return d.value
}

func (d ReleaseDate) String() string {
	return d.value.Format("2006-01-02")
}
