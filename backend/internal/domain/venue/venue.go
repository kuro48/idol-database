package venue

import (
	"errors"
	"time"

	"github.com/kuro48/idol-api/internal/shared/source"
)

const maxNameLength = 200

// Venue は会場を表すエンティティ（集約ルート）
type Venue struct {
	id          VenueID
	name        string
	nameEn      *string
	prefecture  *string
	city        *string
	address     *string
	capacity    *int
	officialURL *string
	sources     []source.Source
	createdAt   time.Time
	updatedAt   time.Time
}

// NewVenue は新しい Venue を生成する
func NewVenue(name string) (*Venue, error) {
	if name == "" {
		return nil, errors.New("会場名は必須です")
	}
	if len([]rune(name)) > maxNameLength {
		return nil, errors.New("会場名は200文字以内で入力してください")
	}
	now := time.Now()
	return &Venue{
		name:      name,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// Reconstruct は永続化されたデータからドメインモデルを再構築する
func Reconstruct(
	id VenueID,
	name string,
	nameEn *string,
	prefecture *string,
	city *string,
	address *string,
	capacity *int,
	officialURL *string,
	sources []source.Source,
	createdAt, updatedAt time.Time,
) *Venue {
	return &Venue{
		id:          id,
		name:        name,
		nameEn:      nameEn,
		prefecture:  prefecture,
		city:        city,
		address:     address,
		capacity:    capacity,
		officialURL: officialURL,
		sources:     sources,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// ---- アクセサ ----

func (v *Venue) ID() VenueID          { return v.id }
func (v *Venue) Name() string         { return v.name }
func (v *Venue) NameEn() *string      { return v.nameEn }
func (v *Venue) Prefecture() *string  { return v.prefecture }
func (v *Venue) City() *string        { return v.city }
func (v *Venue) Address() *string     { return v.address }
func (v *Venue) Capacity() *int       { return v.capacity }
func (v *Venue) OfficialURL() *string { return v.officialURL }
func (v *Venue) CreatedAt() time.Time { return v.createdAt }
func (v *Venue) UpdatedAt() time.Time { return v.updatedAt }

func (v *Venue) Sources() []source.Source {
	if v.sources == nil {
		return []source.Source{}
	}
	return v.sources
}

// ---- ミュータ（updatedAt を自動更新）----

func (v *Venue) SetID(id VenueID) {
	v.id = id
}

func (v *Venue) UpdateName(name string) error {
	if name == "" {
		return errors.New("会場名は必須です")
	}
	if len([]rune(name)) > maxNameLength {
		return errors.New("会場名は200文字以内で入力してください")
	}
	v.name = name
	v.updatedAt = time.Now()
	return nil
}

func (v *Venue) UpdateNameEn(nameEn *string) {
	v.nameEn = nameEn
	v.updatedAt = time.Now()
}

func (v *Venue) UpdatePrefecture(prefecture *string) {
	v.prefecture = prefecture
	v.updatedAt = time.Now()
}

func (v *Venue) UpdateCity(city *string) {
	v.city = city
	v.updatedAt = time.Now()
}

func (v *Venue) UpdateAddress(address *string) {
	v.address = address
	v.updatedAt = time.Now()
}

func (v *Venue) UpdateCapacity(capacity *int) {
	v.capacity = capacity
	v.updatedAt = time.Now()
}

func (v *Venue) UpdateOfficialURL(url *string) {
	v.officialURL = url
	v.updatedAt = time.Now()
}

func (v *Venue) SetSources(sources []source.Source) {
	v.sources = sources
	v.updatedAt = time.Now()
}
