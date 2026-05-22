package group

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListGroupQueryValidateRejectsUnsafeSortAndOrder(t *testing.T) {
	sort := "$where"
	order := "sideways"
	query := ListGroupQuery{Sort: &sort, Order: &order}

	err := query.Validate()

	assert.Error(t, err)
}

func TestListGroupQueryValidateAllowsKnownSortAndOrder(t *testing.T) {
	sort := "formation_date"
	order := "asc"
	query := ListGroupQuery{Sort: &sort, Order: &order}

	err := query.Validate()

	assert.NoError(t, err)
}
