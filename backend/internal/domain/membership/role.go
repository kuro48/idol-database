package membership

import "errors"

type Role string

const (
	RoleMember    Role = "member"
	RoleLeader    Role = "leader"
	RoleSubLeader Role = "sub_leader"
	RoleTrainee   Role = "trainee"
	RoleSupport   Role = "support"
)

func NewRole(s string) (Role, error) {
	r := Role(s)
	if !r.IsValid() {
		return "", errors.New("無効なメンバーシップロールです: " + s)
	}
	return r, nil
}

func (r Role) IsValid() bool {
	switch r {
	case RoleMember, RoleLeader, RoleSubLeader, RoleTrainee, RoleSupport:
		return true
	}
	return false
}

func (r Role) String() string {
	return string(r)
}
