package group

import (
	"errors"
	"time"

	"github.com/kuro48/idol-api/internal/shared/source"
)

type Group struct {
	id            GroupID
	name          GroupName
	status        GroupStatus
	formationDate *FormationDate
	disbandDate   *DisbandDate
	agencyID      *string
	logoURL       *string
	externalIDs   *ExternalIDs
	sources       []source.Source
	createdAt     time.Time
	updatedAt     time.Time
}

func NewGroup(
	name GroupName,
	formationDate *FormationDate,
) (*Group, error) {
	now := time.Now()

	return &Group{
		name:          name,
		formationDate: formationDate,
		disbandDate:   nil,
		status:        GroupStatusActive,
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

func Reconstruct(
	id GroupID,
	name GroupName,
	status GroupStatus,
	formationDate *FormationDate,
	disbandDate *DisbandDate,
	agencyID *string,
	logoURL *string,
	externalIDs *ExternalIDs,
	sources []source.Source,
	createdAt time.Time,
	updatedAt time.Time,
) *Group {
	return &Group{
		id:            id,
		name:          name,
		status:        status,
		formationDate: formationDate,
		disbandDate:   disbandDate,
		agencyID:      agencyID,
		logoURL:       logoURL,
		externalIDs:   externalIDs,
		sources:       sources,
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

func (g *Group) Status() GroupStatus {
	return g.status
}

func (g *Group) AgencyID() *string {
	return g.agencyID
}

func (g *Group) LogoURL() *string {
	return g.logoURL
}

func (g *Group) SetLogoURL(url *string) {
	g.logoURL = url
	g.updatedAt = time.Now()
}

func (g *Group) ExternalIDs() *ExternalIDs {
	if g.externalIDs == nil {
		return NewExternalIDs()
	}
	return g.externalIDs
}

func (g *Group) Sources() []source.Source {
	if g.sources == nil {
		return []source.Source{}
	}
	return g.sources
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

// UpdateStatus は活動状態を更新する
func (g *Group) UpdateStatus(status GroupStatus) error {
	if !status.IsValid() {
		return errors.New("無効なステータスです")
	}
	g.status = status
	g.updatedAt = time.Now()
	return nil
}

// UpdateAgency は所属事務所を更新する
func (g *Group) UpdateAgency(agencyID *string) {
	g.agencyID = agencyID
	g.updatedAt = time.Now()
}

// UpdateExternalIDs は外部IDマッピングを更新する
func (g *Group) UpdateExternalIDs(ids *ExternalIDs) {
	g.externalIDs = ids
	g.updatedAt = time.Now()
}

// SetSources は出典リストを設定する
func (g *Group) SetSources(sources []source.Source) {
	g.sources = sources
	g.updatedAt = time.Now()
}
