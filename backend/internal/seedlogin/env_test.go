package seedlogin

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldApply(t *testing.T) {
	t.Parallel()

	tests := []struct {
		env  string
		want bool
	}{
		{env: "development", want: true},
		{env: "local", want: true},
		{env: "dev", want: true},
		{env: "test", want: true},
		{env: "staging", want: true},
		{env: "STAGING", want: true},
		{env: " production ", want: false},
		{env: "production", want: false},
		{env: "", want: false},
		{env: "preview", want: false},
		{env: "prod", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, ShouldApply(tt.env))
		})
	}
}

func TestAcceptSharedPassword(t *testing.T) {
	t.Parallel()

	demoEmail := Catalog()[0].Email

	assert.True(t, AcceptSharedPassword("staging", demoEmail, SharedPassword))
	assert.True(t, AcceptSharedPassword("development", demoEmail, SharedPassword))
	assert.True(t, AcceptSharedPassword("staging", strings.ToUpper(demoEmail), SharedPassword))
	assert.False(t, AcceptSharedPassword("production", demoEmail, SharedPassword))
	assert.False(t, AcceptSharedPassword("", demoEmail, SharedPassword))
	assert.False(t, AcceptSharedPassword("staging", "stg-operator@example.test", SharedPassword))
	assert.False(t, AcceptSharedPassword("staging", demoEmail, "other-pass"))
	assert.False(t, AcceptSharedPassword("staging", "user@test.com", SharedPassword))
}
