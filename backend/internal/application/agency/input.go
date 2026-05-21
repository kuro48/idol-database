package agency

// CreateInput は事務所作成の入力
type CreateInput struct {
	Name            string
	NameEn          *string
	FoundedDate     *string
	Country         string
	OfficialWebsite *string
	Description     *string
	LogoURL         *string
}

// UpdateInput は事務所更新の入力
type UpdateInput struct {
	ID              string
	Name            *string
	NameEn          *string
	FoundedDate     *string
	OfficialWebsite *string
	Description     *string
	LogoURL         *string
}
