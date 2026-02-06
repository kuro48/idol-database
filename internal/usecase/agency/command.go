package agency

// CreateAgencyCommand は事務所作成コマンド
type CreateAgencyCommand struct {
	Name            string  `json:"name" binding:"required"`
	NameEn          *string `json:"name_en"`
	FoundedDate     *string `json:"founded_date"` // YYYY-MM-DD
	Country         string  `json:"country" binding:"required"`
	OfficialWebsite *string `json:"official_website"`
	Description     *string `json:"description"`
	LogoURL         *string `json:"logo_url"`
}

// UpdateAgencyCommand は事務所更新コマンド
type UpdateAgencyCommand struct {
	ID              string  `json:"-"`
	Name            *string `json:"name"`
	NameEn          *string `json:"name_en"`
	FoundedDate     *string `json:"founded_date"` // YYYY-MM-DD
	OfficialWebsite *string `json:"official_website"`
	Description     *string `json:"description"`
	LogoURL         *string `json:"logo_url"`
}

// DeleteAgencyCommand は事務所削除コマンド
type DeleteAgencyCommand struct {
	ID string
}
