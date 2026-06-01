package idol

import (
	"errors"
	"time"

	"github.com/kuro48/idol-api/internal/shared/source"
)

// Idol はアイドル集約のルートエンティティ
type Idol struct {
	id          IdolID
	name        IdolName
	birthdate   *Birthdate
	status      IdolStatus
	agencyID    *string      // 所属事務所ID（Phase2でMembershipに移行予定）
	socialLinks *SocialLinks
	externalIDs *ExternalIDs
	profileImageURL *string
	tagIDs          []string
	aliases         []string
	sources         []source.Source
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
		name:      name,
		birthdate: birthdate,
		status:    IdolStatusActive,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// Reconstruct はデータストアからアイドルを再構築する（永続化層用）
func Reconstruct(
	id IdolID,
	name IdolName,
	birthdate *Birthdate,
	status IdolStatus,
	agencyID *string,
	profileImageURL *string,
	socialLinks *SocialLinks,
	externalIDs *ExternalIDs,
	tagIDs []string,
	aliases []string,
	sources []source.Source,
	createdAt time.Time,
	updatedAt time.Time,
) *Idol {
	return &Idol{
		id:              id,
		name:            name,
		birthdate:       birthdate,
		status:          status,
		agencyID:        agencyID,
		profileImageURL: profileImageURL,
		socialLinks:     socialLinks,
		externalIDs:     externalIDs,
		tagIDs:          tagIDs,
		aliases:         aliases,
		sources:         sources,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
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

func (i *Idol) ProfileImageURL() *string {
	return i.profileImageURL
}

func (i *Idol) SocialLinks() *SocialLinks {
	return i.socialLinks
}

func (i *Idol) ExternalIDs() *ExternalIDs {
	if i.externalIDs == nil {
		return NewExternalIDs()
	}
	return i.externalIDs
}

func (i *Idol) TagIDs() []string {
	return i.tagIDs
}

func (i *Idol) Aliases() []string {
	if i.aliases == nil {
		return []string{}
	}
	return i.aliases
}

func (i *Idol) Status() IdolStatus {
	return i.status
}

func (i *Idol) Sources() []source.Source {
	if i.sources == nil {
		return []source.Source{}
	}
	return i.sources
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

// SetProfileImageURL はプロフィール画像URLを設定する
func (i *Idol) SetProfileImageURL(url *string) {
	i.profileImageURL = url
	i.updatedAt = time.Now()
}

// UpdateSocialLinks はSNS/外部リンクを更新する
func (i *Idol) UpdateSocialLinks(links *SocialLinks) {
	i.socialLinks = links
	i.updatedAt = time.Now()
}

// UpdateExternalIDs は外部IDマッピングを更新する
func (i *Idol) UpdateExternalIDs(ids *ExternalIDs) {
	i.externalIDs = ids
	i.updatedAt = time.Now()
}

// AddTag はタグを追加する（重複チェックあり）
func (i *Idol) AddTag(tagID string) {
	// 既に存在する場合は追加しない
	for _, existingID := range i.tagIDs {
		if existingID == tagID {
			return
		}
	}
	i.tagIDs = append(i.tagIDs, tagID)
	i.updatedAt = time.Now()
}

// RemoveTag はタグを削除する
func (i *Idol) RemoveTag(tagID string) {
	newTagIDs := make([]string, 0, len(i.tagIDs))
	for _, id := range i.tagIDs {
		if id != tagID {
			newTagIDs = append(newTagIDs, id)
		}
	}
	i.tagIDs = newTagIDs
	i.updatedAt = time.Now()
}

// SetTags はタグIDのリストを設定する
func (i *Idol) SetTags(tagIDs []string) {
	i.tagIDs = tagIDs
	i.updatedAt = time.Now()
}

// SetAliases は別名一覧を設定する
func (i *Idol) SetAliases(aliases []string) {
	i.aliases = aliases
	i.updatedAt = time.Now()
}

// UpdateStatus は活動状態を更新する
func (i *Idol) UpdateStatus(status IdolStatus) error {
	if !status.IsValid() {
		return errors.New("無効なステータスです")
	}
	i.status = status
	i.updatedAt = time.Now()
	return nil
}

// SetSources は出典リストを設定する
func (i *Idol) SetSources(sources []source.Source) {
	i.sources = sources
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
