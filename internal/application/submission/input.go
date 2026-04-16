package submission

// CreateInput は投稿審査作成の入力
type CreateInput struct {
	TargetType       string
	Payload          string
	SourceURLs       []string
	ContributorEmail string
}

// UpdateStatusInput はステータス更新の入力
type UpdateStatusInput struct {
	Status       string // approved/rejected/needs_revision
	ReviewedBy   string
	RevisionNote string // needs_revision 時のみ
}
