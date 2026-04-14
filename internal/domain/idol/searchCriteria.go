package idol

import "time"

type SearchCriteria struct {
	Name          *string
	AgencyID      *string
	AgeMin        *int
	AgeMax        *int
	BirthdateFrom *time.Time
	BirthdateTo   *time.Time

	Sort  string
	Order string

	Offset int
	Limit  int
}
