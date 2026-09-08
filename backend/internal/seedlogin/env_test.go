package seedlogin

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestOperatorFromEnv(t *testing.T) {
	t.Setenv(operatorEnvEmail, "")
	t.Setenv(operatorEnvName, "")
	t.Setenv(operatorEnvPassword, "")
	spec, ok, err := operatorFromEnv()
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, spec.Email)

	t.Setenv(operatorEnvEmail, "stg-operator@example.test")
	_, _, err = operatorFromEnv()
	require.Error(t, err)
	assert.Contains(t, err.Error(), operatorEnvName)

	t.Setenv(operatorEnvName, "Operator")
	t.Setenv(operatorEnvPassword, "OperatorPass1")
	spec, ok, err = operatorFromEnv()
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "stg-operator@example.test", spec.Email)
	assert.Equal(t, "Operator", spec.Name)

	t.Setenv(operatorEnvEmail, Catalog()[0].Email)
	_, _, err = operatorFromEnv()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "demo catalog")

	t.Setenv(operatorEnvEmail, "stg-operator@example.test")
	t.Setenv(operatorEnvPassword, "short")
	_, _, err = operatorFromEnv()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "8 characters")
}
