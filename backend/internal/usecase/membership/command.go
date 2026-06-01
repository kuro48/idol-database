package membership

type CreateMembershipCommand struct {
	IdolID   string
	GroupID  string
	Role     string
	JoinedAt *string
}

type UpdateMembershipCommand struct {
	ID       string
	Role     *string
	JoinedAt *string
	LeftAt   *string
}

type DeleteMembershipCommand struct {
	ID string
}
