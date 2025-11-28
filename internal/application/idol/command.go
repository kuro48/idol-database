package idol

// CreateIdolCommand はアイドル作成コマンド
type CreateIdolCommand struct {
	Name        string
	Birthdate   *string
	AgencyID    *string
}

// UpdateIdolCommand はアイドル更新コマンド
type UpdateIdolCommand struct {
	ID          string
	Name        *string
	Birthdate   *string
	AgencyID    *string
}

// DeleteIdolCommand はアイドル削除コマンド
type DeleteIdolCommand struct {
	ID string
}
