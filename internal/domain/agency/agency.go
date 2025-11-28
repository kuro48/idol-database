package agency

import "time"

// Agency は事務所エンティティ
type Agency struct {
	id              AgencyID
	name            AgencyName
	nameEn          *string    // 英語名（オプション）
	foundedDate     *time.Time // 設立日（オプション）
	country         Country
	officialWebsite *string // 公式サイトURL
	description     *string // 説明
	logoURL         *string // ロゴ画像URL
	createdAt       time.Time
	updatedAt       time.Time
}

// NewAgency は新しい事務所を作成する
func NewAgency(
	id AgencyID,
	name AgencyName,
	country Country,
) *Agency {
	now := time.Now()
	return &Agency{
		id:        id,
		name:      name,
		country:   country,
		createdAt: now,
		updatedAt: now,
	}
}

// Getters
func (a *Agency) ID() AgencyID              { return a.id }
func (a *Agency) Name() AgencyName          { return a.name }
func (a *Agency) NameEn() *string           { return a.nameEn }
func (a *Agency) FoundedDate() *time.Time   { return a.foundedDate }
func (a *Agency) Country() Country          { return a.country }
func (a *Agency) OfficialWebsite() *string  { return a.officialWebsite }
func (a *Agency) Description() *string      { return a.description }
func (a *Agency) LogoURL() *string          { return a.logoURL }
func (a *Agency) CreatedAt() time.Time      { return a.createdAt }
func (a *Agency) UpdatedAt() time.Time      { return a.updatedAt }

// UpdateDetails は事務所の詳細情報を更新する
func (a *Agency) UpdateDetails(
	name *AgencyName,
	nameEn *string,
	foundedDate *time.Time,
	officialWebsite *string,
	description *string,
	logoURL *string,
) {
	if name != nil {
		a.name = *name
	}
	a.nameEn = nameEn
	a.foundedDate = foundedDate
	a.officialWebsite = officialWebsite
	a.description = description
	a.logoURL = logoURL
	a.updatedAt = time.Now()
}
