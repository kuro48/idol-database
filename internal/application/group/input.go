package group

// CreateInput はグループ作成の入力
type CreateInput struct {
	Name          string
	FormationDate *string
	DisbandDate   *string
}

// UpdateInput はグループ更新の入力
type UpdateInput struct {
	ID            string
	Name          *string
	FormationDate *string
	DisbandDate   *string
}
