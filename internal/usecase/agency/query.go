package agency

// GetAgencyQuery は事務所取得クエリ
type GetAgencyQuery struct {
	ID      string   `uri:"id" binding:"required"`
	Include []string `form:"include"` // (未実装) idols, groups
}

// ListAgenciesQuery は事務所一覧取得クエリ
type ListAgenciesQuery struct {
	Include []string `form:"include"` // (未実装) idols
}

// AgencyDTO は事務所のデータ転送オブジェクト
type AgencyDTO struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	NameEn          *string `json:"name_en,omitempty"`
	FoundedDate     *string `json:"founded_date,omitempty"` // YYYY-MM-DD
	Country         string  `json:"country"`
	OfficialWebsite *string `json:"official_website,omitempty"`
	Description     *string `json:"description,omitempty"`
	LogoURL         *string `json:"logo_url,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}
