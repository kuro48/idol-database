package membership

type CreateInput struct {
	IdolID   string
	GroupID  string
	Role     string
	JoinedAt *string // "2006-01-02" or nil
}

type UpdateInput struct {
	ID       string
	Role     *string
	JoinedAt *string // "2006-01-02" or nil; omitted means no change
	LeftAt   *string // "2006-01-02" or nil; empty string means clear
}
