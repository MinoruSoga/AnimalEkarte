package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthCompositionDoesNotWrapCurrentAccessCache(t *testing.T) {
	source, err := os.ReadFile("composition_auth.go")
	require.NoError(t, err)
	assert.NotContains(t, string(source), "NewCachedCurrentAccessResolver")
	abs, err := filepath.Abs("composition_auth.go")
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(abs, "composition_auth.go"))
}
