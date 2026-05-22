package agency

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListAgenciesQueryValidateRejectsUnsafeSortAndOrder(t *testing.T) {
	sort := "$where"
	order := "sideways"
	query := ListAgenciesQuery{Sort: &sort, Order: &order}

	err := query.Validate()

	assert.Error(t, err)
}

func TestListAgenciesQueryValidateAllowsKnownSortAndOrder(t *testing.T) {
	sort := "founded_date"
	order := "asc"
	query := ListAgenciesQuery{Sort: &sort, Order: &order}

	err := query.Validate()

	assert.NoError(t, err)
}
