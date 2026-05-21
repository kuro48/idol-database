package release

import "errors"

// ReleaseID はリリースの一意識別子
type ReleaseID struct {
	value string
}

func NewReleaseID(value string) (ReleaseID, error) {
	if value == "" {
		return ReleaseID{}, errors.New("リリースIDは空にできません")
	}
	return ReleaseID{value: value}, nil
}

func (id ReleaseID) Value() string {
	return id.value
}

func (id ReleaseID) Equals(other ReleaseID) bool {
	return id.value == other.value
}
