package group

import "errors"

// GroupStatus はグループの活動状態
type GroupStatus string

const (
	GroupStatusActive    GroupStatus = "active"
	GroupStatusDisbanded GroupStatus = "disbanded"
	GroupStatusHiatus    GroupStatus = "hiatus"
)

func NewGroupStatus(s string) (GroupStatus, error) {
	switch GroupStatus(s) {
	case GroupStatusActive, GroupStatusDisbanded, GroupStatusHiatus:
		return GroupStatus(s), nil
	}
	return "", errors.New("無効なグループステータスです: active / disbanded / hiatus のいずれかを指定してください")
}

func (s GroupStatus) IsValid() bool {
	_, err := NewGroupStatus(string(s))
	return err == nil
}
