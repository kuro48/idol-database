package venue

// CreateInput は会場作成の入力データ
type CreateInput struct {
	Name        string
	NameEn      *string
	Prefecture  *string
	City        *string
	Address     *string
	Capacity    *int
	OfficialURL *string
}

// UpdateInput は会場更新の入力データ
type UpdateInput struct {
	ID          string
	Name        *string
	NameEn      *string
	Prefecture  *string
	City        *string
	Address     *string
	Capacity    *int
	OfficialURL *string
}
