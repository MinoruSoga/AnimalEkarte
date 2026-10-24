package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appcrypto "github.com/animal-ekarte/backend/internal/infra/crypto"
)

// 予約の完了・キャンセル Push のトークン復号は、lstep 連携設定と同じ cipher・同じ規則で行う。
// 復号失敗は空文字へ置換せず error にする（呼び出し側が Push を止めてログに残す）。
func TestReservationLineTokenDecryptor(t *testing.T) {
	newCipher := func(t *testing.T) *appcrypto.AESGCMCipher {
		t.Helper()
		cipher, err := appcrypto.NewAESGCMCipher(strings.Repeat("ab", 32))
		require.NoError(t, err)
		return cipher
	}

	t.Run("decrypts a value stored by the integration cipher", func(t *testing.T) {
		cipher := newCipher(t)
		stored, err := cipher.Encrypt("token-bbbb")
		require.NoError(t, err)

		got, err := reservationLineTokenDecryptor(cipher)(stored)

		require.NoError(t, err)
		assert.Equal(t, "token-bbbb", got)
	})

	t.Run("a nil cipher (development) passes the stored value through", func(t *testing.T) {
		got, err := reservationLineTokenDecryptor(nil)("token-bbbb")

		require.NoError(t, err)
		assert.Equal(t, "token-bbbb", got)
	})

	t.Run("a value that cannot be decrypted is an error, never an empty token", func(t *testing.T) {
		got, err := reservationLineTokenDecryptor(newCipher(t))("not-ciphertext")

		require.Error(t, err)
		assert.Empty(t, got)
	})
}
