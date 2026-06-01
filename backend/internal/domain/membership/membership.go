package membership

import (
	"errors"
	"time"

	"github.com/kuro48/idol-api/internal/shared/source"
)

type Membership struct {
	id        MembershipID
	idolID    string
	groupID   string
	role      Role
	joinedAt  *time.Time
	leftAt    *time.Time
	sources   []source.Source
	createdAt time.Time
	updatedAt time.Time
}

func NewMembership(idolID, groupID string, role Role, joinedAt *time.Time) (*Membership, error) {
	if idolID == "" {
		return nil, errors.New("アイドルIDは必須です")
	}
	if groupID == "" {
		return nil, errors.New("グループIDは必須です")
	}
	now := time.Now()
	return &Membership{
		idolID:    idolID,
		groupID:   groupID,
		role:      role,
		joinedAt:  joinedAt,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func Reconstruct(
	id MembershipID,
	idolID, groupID string,
	role Role,
	joinedAt, leftAt *time.Time,
	sources []source.Source,
	createdAt, updatedAt time.Time,
) *Membership {
	return &Membership{
		id:        id,
		idolID:    idolID,
		groupID:   groupID,
		role:      role,
		joinedAt:  joinedAt,
		leftAt:    leftAt,
		sources:   sources,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (m *Membership) ID() MembershipID     { return m.id }
func (m *Membership) IdolID() string       { return m.idolID }
func (m *Membership) GroupID() string      { return m.groupID }
func (m *Membership) Role() Role           { return m.role }
func (m *Membership) JoinedAt() *time.Time { return m.joinedAt }
func (m *Membership) LeftAt() *time.Time   { return m.leftAt }
func (m *Membership) IsActive() bool       { return m.leftAt == nil }
func (m *Membership) CreatedAt() time.Time { return m.createdAt }
func (m *Membership) UpdatedAt() time.Time { return m.updatedAt }
func (m *Membership) Sources() []source.Source {
	if m.sources == nil {
		return []source.Source{}
	}
	return m.sources
}

func (m *Membership) SetID(id MembershipID) {
	m.id = id
}

func (m *Membership) UpdateRole(role Role) error {
	if !role.IsValid() {
		return errors.New("無効なロールです")
	}
	m.role = role
	m.updatedAt = time.Now()
	return nil
}

func (m *Membership) UpdateJoinedAt(t *time.Time) {
	m.joinedAt = t
	m.updatedAt = time.Now()
}

func (m *Membership) Leave(leftAt time.Time) error {
	if m.leftAt != nil {
		return errors.New("既に脱退済みです")
	}
	if m.joinedAt != nil && leftAt.Before(*m.joinedAt) {
		return errors.New("脱退日は加入日より後である必要があります")
	}
	m.leftAt = &leftAt
	m.updatedAt = time.Now()
	return nil
}

func (m *Membership) ClearLeftAt() {
	m.leftAt = nil
	m.updatedAt = time.Now()
}

func (m *Membership) SetSources(sources []source.Source) {
	m.sources = sources
	m.updatedAt = time.Now()
}
