package venue

import "errors"

// VenueID は会場の識別子を表す値オブジェクト
type VenueID struct {
	value string
}

func NewVenueID(value string) (VenueID, error) {
	if value == "" {
		return VenueID{}, errors.New("会場IDは空にできません")
	}
	return VenueID{value: value}, nil
}

func (id VenueID) Value() string {
	return id.value
}

func (id VenueID) Equals(other VenueID) bool {
	return id.value == other.value
}
