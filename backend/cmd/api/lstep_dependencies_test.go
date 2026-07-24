package main

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appcrypto "github.com/animal-ekarte/backend/internal/infra/crypto"
)

func TestReservationCredentialAndMessengerWiring(t *testing.T) {
	cipher, err := appcrypto.NewAESGCMCipher(strings.Repeat("01", 32))
	require.NoError(t, err)

	encrypt := reservationCredentialEncryptor(cipher)
	decrypt := reservationCredentialDecryptor(cipher)
	encrypted, err := encrypt("plain-value")
	require.NoError(t, err)
	assert.NotEqual(t, "plain-value", encrypted)
	assert.Equal(t, "plain-value", decrypt(context.Background(), encrypted))
	assert.NotNil(t, newReservationLineMessenger("channel-token"))
}
