package submission

// CreateSubmissionCommand は投稿審査作成コマンド
type CreateSubmissionCommand struct {
	TargetType       string
	Payload          map[string]interface{}
	SourceURLs       []string
	ContributorEmail string
}

// UpdateStatusCommand はステータス更新コマンド
type UpdateStatusCommand struct {
	ID           string
	Status       string
	ReviewedBy   string
	RevisionNote string
}

// ReviseSubmissionCommand は再投稿コマンド
type ReviseSubmissionCommand struct {
	ID         string
	Payload    map[string]interface{}
	SourceURLs []string
}
