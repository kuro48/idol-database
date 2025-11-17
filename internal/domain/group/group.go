package group

import (
	"errors"
	"time"
)

type Group struct {
	id            GroupID
	name          GroupName
	formationDate *FormationDate
	disbandDate   *DisbandDate
	createdAt     time.Time
	updatedAt     time.Time
}

func NewGroup(
	name          GroupName,
	formationDate *FormationDate,
) (*Group, error) {
	now := time.Now()

	return &Group{
		name:          name,
		formationDate: formationDate,
		disbandDate:   nil, // 新規作成時は未解散
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

func Reconstruct(
	id            GroupID,
	name          GroupName,
	formationDate *FormationDate,
	disbandDate   *DisbandDate,
	createdAt     time.Time,
	updatedAt     time.Time,
) *Group {
	return &Group{
		id:            id,
		name:          name,
		formationDate: formationDate,
		disbandDate:   disbandDate,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

func (g *Group) ID() GroupID {
	return g.id
}

func (g *Group) Name() GroupName {
	return g.name
}

func (g *Group) FormationDate() *FormationDate {
	return g.formationDate
}

func (g *Group) DisbandDate() *DisbandDate {
	return g.disbandDate
}

func (g *Group) CreatedAt() time.Time {
	return g.createdAt
}

func (g *Group) UpdatedAt() time.Time {
	return g.updatedAt
}

func (g *Group) SetID(id GroupID) {
	g.id = id
}

func (g *Group) ChangeName(name GroupName) error {
	if name.Value() == "" {
		return errors.New("名前は空にできません")
	}
	g.name = name
	g.updatedAt = time.Now()
	return nil
}

func (g *Group) UpdateFormationDate(formationDate FormationDate) error {
	g.formationDate = &formationDate
	g.updatedAt = time.Now()
	return nil
}

func (g *Group) UpdateDisbandDate(disbandDate DisbandDate) error {
	g.disbandDate = &disbandDate
	g.updatedAt = time.Now()
	return nil
}

func (g *Group) Disband(disbandDate DisbandDate) error {
	// ✅ 結成日が登録されている場合、解散日は結成日より後
	if g.formationDate != nil && disbandDate.Value().Before(g.formationDate.Value()) {
		return errors.New("解散日は結成日より後である必要があります")
	}

    // ✅ 既に解散している場合はエラー
    if g.disbandDate != nil {
		return errors.New("既に解散済みのグループです")
	}

    g.disbandDate = &disbandDate
    g.updatedAt = time.Now()
    return nil
}

func (g *Group) IsActive() bool {
	return g.disbandDate == nil
}