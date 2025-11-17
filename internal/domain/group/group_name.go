package group

import (
	"errors"
)

// GroupName はグループ名
type GroupName struct {
	value string
}

func NewGroupName(value string) (GroupName, error) {
	if value == "" {
		// グループなしは許可
		return GroupName{value: ""}, nil
	}
	if len(value) > 100 {
		return GroupName{}, errors.New("グループ名は100文字以内である必要があります")
	}
	return GroupName{value: value}, nil
}

func (g GroupName) Value() string {
	return g.value
}

func (g GroupName) IsEmpty() bool {
	return g.value == ""
}