package mongodb

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafePartialMatchRegexEscapesUserInput(t *testing.T) {
	pattern := safePartialMatchRegex("Tokyo.*(idol)?")

	assert.Equal(t, regexp.QuoteMeta("Tokyo.*(idol)?"), pattern)
	assert.Regexp(t, pattern, "Tokyo.*(idol)?")
	assert.NotRegexp(t, pattern, "Tokyo super idol")
}
