package source

import (
	"errors"
	"strings"
	"time"
)

// Type は出典の種別
type Type string

const (
	TypeOfficial Type = "official"
	TypeNews     Type = "news"
	TypeWiki     Type = "wiki"
	TypeFan      Type = "fan"
	TypeOther    Type = "other"
)

func (t Type) IsValid() bool {
	switch t {
	case TypeOfficial, TypeNews, TypeWiki, TypeFan, TypeOther:
		return true
	}
	return false
}

// Source は情報の根拠となる出典を表す値オブジェクト
type Source struct {
	url        string
	sourceType Type
	isPrimary  bool
	notedAt    time.Time
}

// New は出典を生成する
func New(url string, sourceType Type, isPrimary bool, notedAt time.Time) (Source, error) {
	if url == "" {
		return Source{}, errors.New("出典URLは空にできません")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return Source{}, errors.New("URLはhttp://またはhttps://で始まる必要があります")
	}
	if !sourceType.IsValid() {
		return Source{}, errors.New("無効な出典種別です: official / news / wiki / fan / other のいずれかを指定してください")
	}
	return Source{
		url:        url,
		sourceType: sourceType,
		isPrimary:  isPrimary,
		notedAt:    notedAt,
	}, nil
}

// Reconstruct は永続化層から出典を再構築する（バリデーションなし）
func Reconstruct(url string, sourceType Type, isPrimary bool, notedAt time.Time) Source {
	return Source{
		url:        url,
		sourceType: sourceType,
		isPrimary:  isPrimary,
		notedAt:    notedAt,
	}
}

func (s Source) URL() string       { return s.url }
func (s Source) Type() Type        { return s.sourceType }
func (s Source) IsPrimary() bool   { return s.isPrimary }
func (s Source) NotedAt() time.Time { return s.notedAt }
