package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeRawQueryMasksSensitiveValues(t *testing.T) {
	got := sanitizeRawQuery("email=user@example.com&access_token=secret-token&name=idol")

	assert.Contains(t, got, "email=%5BREDACTED%5D")
	assert.Contains(t, got, "access_token=%5BREDACTED%5D")
	assert.Contains(t, got, "name=idol")
	assert.NotContains(t, got, "user@example.com")
	assert.NotContains(t, got, "secret-token")
}

func TestSanitizeRawQueryPreservesMalformedQueryWithoutValues(t *testing.T) {
	got := sanitizeRawQuery("%%%")

	assert.Equal(t, "[invalid-query]", got)
}
