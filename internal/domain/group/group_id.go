package group

import (
	"errors"
)

type GroupID struct {
	value string
}

func NewGroupID(value string) (GroupID, error) {
	if value == "" {
		return GroupID{}, errors.New("グループIDはからにできません")
	}
	return GroupID{value: value}, nil
}

func (id GroupID) Value() string {
	return id.value
}

func (id GroupID) Equals(other GroupID) bool {
	return id.value == other.value
}