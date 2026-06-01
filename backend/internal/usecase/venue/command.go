package venue

// CreateVenueCommand は会場作成コマンド
type CreateVenueCommand struct {
	Name        string
	NameEn      *string
	Prefecture  *string
	City        *string
	Address     *string
	Capacity    *int
	OfficialURL *string
}

// UpdateVenueCommand は会場更新コマンド
type UpdateVenueCommand struct {
	ID          string
	Name        *string
	NameEn      *string
	Prefecture  *string
	City        *string
	Address     *string
	Capacity    *int
	OfficialURL *string
}

// DeleteVenueCommand は会場削除コマンド
type DeleteVenueCommand struct {
	ID string
}
