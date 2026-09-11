package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_RejectsUnknownCommand(t *testing.T) {
	err := run([]string{"explode"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "unknown command")
}

func TestRun_RequiresUsage(t *testing.T) {
	err := run(nil)
	require.Error(t, err)
	assert.ErrorContains(t, err, "usage")
}
