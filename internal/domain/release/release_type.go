package release

import "fmt"

// ReleaseType はリリース種別
type ReleaseType string

const (
	ReleaseTypeSingle        ReleaseType = "single"
	ReleaseTypeAlbum         ReleaseType = "album"
	ReleaseTypeEP            ReleaseType = "ep"
	ReleaseTypeMiniAlbum     ReleaseType = "mini_album"
	ReleaseTypeDigitalSingle ReleaseType = "digital_single"
	ReleaseTypeCompilation   ReleaseType = "compilation"
)

var validReleaseTypes = map[ReleaseType]struct{}{
	ReleaseTypeSingle:        {},
	ReleaseTypeAlbum:         {},
	ReleaseTypeEP:            {},
	ReleaseTypeMiniAlbum:     {},
	ReleaseTypeDigitalSingle: {},
	ReleaseTypeCompilation:   {},
}

func NewReleaseType(value string) (ReleaseType, error) {
	t := ReleaseType(value)
	if _, ok := validReleaseTypes[t]; !ok {
		return "", fmt.Errorf("無効なリリース種別です: %s (single/album/ep/mini_album/digital_single/compilation)", value)
	}
	return t, nil
}

func (t ReleaseType) IsValid() bool {
	_, ok := validReleaseTypes[t]
	return ok
}

func (t ReleaseType) Value() string {
	return string(t)
}

// ValidReleaseTypes は有効なリリース種別の一覧を返す
func ValidReleaseTypes() []ReleaseType {
	types := make([]ReleaseType, 0, len(validReleaseTypes))
	for t := range validReleaseTypes {
		types = append(types, t)
	}
	return types
}
