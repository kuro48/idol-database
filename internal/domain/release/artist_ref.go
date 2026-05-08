package release

import "errors"

// ArtistKind はアーティスト参照の種別（アイドル個人 or グループ）
type ArtistKind string

const (
	ArtistKindIdol  ArtistKind = "idol"
	ArtistKindGroup ArtistKind = "group"
)

// ArtistRef はリリースに関わるアーティストへの参照
type ArtistRef struct {
	kind ArtistKind
	id   string // idol または group の ID
	role string // "center", "feat", "main", "guest" 等（空可）
}

func NewArtistRef(kind ArtistKind, id, role string) (ArtistRef, error) {
	if kind != ArtistKindIdol && kind != ArtistKindGroup {
		return ArtistRef{}, errors.New("アーティスト種別は idol または group である必要があります")
	}
	if id == "" {
		return ArtistRef{}, errors.New("アーティストIDは空にできません")
	}
	return ArtistRef{kind: kind, id: id, role: role}, nil
}

func (a ArtistRef) Kind() ArtistKind {
	return a.kind
}

func (a ArtistRef) ID() string {
	return a.id
}

func (a ArtistRef) Role() string {
	return a.role
}
