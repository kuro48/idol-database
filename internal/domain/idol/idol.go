package idol

import (
	"errors"
	"time"
)

// Idol はアイドル集約のルートエンティティ
type Idol struct {
	id          IdolID
	name        IdolName
	birthdate   *Birthdate
	agencyID    *string    // 所属事務所ID（オプショナル）
	createdAt   time.Time
	updatedAt   time.Time
}

// NewIdol は新しいアイドルを作成する
func NewIdol(
	name IdolName,
	birthdate *Birthdate,
) (*Idol, error) {
	now := time.Now()

	return &Idol{
		name:        name,
		birthdate:   birthdate,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// Reconstruct はデータストアからアイドルを再構築する（永続化層用）
func Reconstruct(
	id IdolID,
	name IdolName,
	birthdate *Birthdate,
	agencyID *string,
	createdAt time.Time,
	updatedAt time.Time,
) *Idol {
	return &Idol{
		id:          id,
		name:        name,
		birthdate:   birthdate,
		agencyID:    agencyID,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// ゲッター

func (i *Idol) ID() IdolID {
	return i.id
}

func (i *Idol) Name() IdolName {
	return i.name
}

func (i *Idol) Birthdate() *Birthdate {
	return i.birthdate
}

func (i *Idol) AgencyID() *string {
	return i.agencyID
}

func (i *Idol) CreatedAt() time.Time {
	return i.createdAt
}

func (i *Idol) UpdatedAt() time.Time {
	return i.updatedAt
}

// ビジネスロジック

// ChangeName はアイドルの名前を変更する
func (i *Idol) ChangeName(name IdolName) error {
	if name.Value() == "" {
		return errors.New("名前は空にできません")
	}
	i.name = name
	i.updatedAt = time.Now()
	return nil
}

// UpdateBirthdate は生年月日を更新する
func (i *Idol) UpdateBirthdate(birthdate *Birthdate) {
	i.birthdate = birthdate
	i.updatedAt = time.Now()
}

// UpdateAgency は所属事務所を更新する
func (i *Idol) UpdateAgency(agencyID *string) {
	i.agencyID = agencyID
	i.updatedAt = time.Now()
}

// SetID はIDを設定する（永続化後に使用）
func (i *Idol) SetID(id IdolID) {
	i.id = id
}

// Age は現在の年齢を返す
func (i *Idol) Age() (int, error) {
	if i.birthdate == nil {
		return 0, errors.New("生年月日が登録されていないため、年齢を計算できません")
	}
	return i.birthdate.Age(), nil
}

// Validate はアイドルの状態が有効かを検証する
func (i *Idol) Validate() error {
	if i.name.Value() == "" {
		return errors.New("名前は必須です")
	}
	return nil
}
