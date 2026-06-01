package membership

import "errors"

type MembershipID struct {
	value string
}

func NewMembershipID(value string) (MembershipID, error) {
	if value == "" {
		return MembershipID{}, errors.New("メンバーシップIDは空にできません")
	}
	return MembershipID{value: value}, nil
}

func (id MembershipID) Value() string {
	return id.value
}

func (id MembershipID) Equals(other MembershipID) bool {
	return id.value == other.value
}
