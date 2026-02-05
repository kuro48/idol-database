package idol

// CreateInput はアイドル作成の入力
// usecase層から渡される前提のため、HTTP由来のタグは持たない
type CreateInput struct {
	Name      string
	Birthdate *string
	AgencyID  *string
}

// UpdateInput はアイドル更新の入力
type UpdateInput struct {
	ID        string
	Name      *string
	Birthdate *string
	AgencyID  *string
}

// UpdateSocialLinksInput はSNS/外部リンク更新の入力
type UpdateSocialLinksInput struct {
	ID              string
	Twitter         *string
	Instagram       *string
	TikTok          *string
	YouTube         *string
	Facebook        *string
	OfficialWebsite *string
	FanClub         *string
}
