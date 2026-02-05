package idol

// CreateIdolCommand はアイドル作成コマンド
type CreateIdolCommand struct {
	Name      string
	Birthdate *string
	AgencyID  *string
}

// UpdateIdolCommand はアイドル更新コマンド
type UpdateIdolCommand struct {
	ID        string
	Name      *string
	Birthdate *string
	AgencyID  *string
}

// DeleteIdolCommand はアイドル削除コマンド
type DeleteIdolCommand struct {
	ID string
}

// UpdateSocialLinksCommand はSNS/外部リンク更新コマンド
type UpdateSocialLinksCommand struct {
	ID              string
	Twitter         *string `json:"twitter"`
	Instagram       *string `json:"instagram"`
	TikTok          *string `json:"tiktok"`
	YouTube         *string `json:"youtube"`
	Facebook        *string `json:"facebook"`
	OfficialWebsite *string `json:"official_website"`
	FanClub         *string `json:"fan_club"`
}
