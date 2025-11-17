package group

type GetGroupQuery struct {
	ID string
}

type ListGroupQuery struct {
	Limit  int
	Offset int
}

type GroupDTO struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	FormationDate string `json:"formation_date,omitempty"`
	DisbandDate   string `json:"disband_date,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}