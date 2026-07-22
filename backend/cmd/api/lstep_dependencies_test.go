package main

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appcrypto "github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/lstep"
)

func TestNewLegacyLstepDependencies_CredentialAndMessengerWiring(t *testing.T) {
	cipher, err := appcrypto.NewAESGCMCipher(strings.Repeat("01", 32))
	require.NoError(t, err)

	deps := newLegacyLstepDependencies(&lstep.Application{}, cipher)
	encrypted, err := deps.EncryptCredential("plain-value")
	require.NoError(t, err)
	assert.NotEqual(t, "plain-value", encrypted)
	assert.Equal(t, "plain-value", deps.DecryptCredential(context.Background(), encrypted))
	assert.NotNil(t, deps.NewLinePusher("channel-token"))
}
