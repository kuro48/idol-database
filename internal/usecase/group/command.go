package group

type CreateGroupCommand struct {
	Name          string
	FormationDate *string
	DisbandDate   *string
}

type UpdateGroupCommand struct {
	ID            string
	Name          *string
	FormationDate *string
	DisbandDate   *string
}

type DeleteGroupCommand struct {
	ID string
}
