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

// UpdateExternalIDsCommand は外部IDマッピング更新コマンド
// ExternalIDs のキーは ExternalIDKind の文字列値（例: "twitter", "youtube_channel"）
// 空文字列を指定した場合はそのIDを削除する
type UpdateExternalIDsCommand struct {
	ID          string
	ExternalIDs map[string]string
}
