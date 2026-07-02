package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/infra/crypto"
)

// TestLineCredentialEncryptDecrypt は LINE 認証情報の暗号化/復号ヘルパーを検証する（H-4）。
// testIntegrationKeyHex は line_reservation_setting_service_test.go で定義。
func TestLineCredentialEncryptDecrypt(t *testing.T) {
	ctx := context.Background()
	cipher, err := crypto.NewAESGCMCipher(testIntegrationKeyHex)
	if err != nil {
		t.Fatalf("failed to build test cipher: %v", err)
	}

	t.Run("round-trip: encrypted value decrypts back to plaintext", func(t *testing.T) {
		enc, err := encryptLineCredential(cipher, "channel-secret-123")
		assert.NoError(t, err)
		assert.NotEqual(t, "channel-secret-123", enc)

		got := decryptLineCredential(ctx, cipher, enc)
		assert.Equal(t, "channel-secret-123", got)
	})

	t.Run("legacy plaintext row falls back to the raw value", func(t *testing.T) {
		// 暗号化導入前に保存された平文（AES-GCM 復号に失敗する）
		got := decryptLineCredential(ctx, cipher, "5344ef84eb7072b5894f7e087db28827")
		assert.Equal(t, "5344ef84eb7072b5894f7e087db28827", got)
	})

	t.Run("nil cipher passes the value through unchanged", func(t *testing.T) {
		enc, err := encryptLineCredential(nil, "raw-value")
		assert.NoError(t, err)
		assert.Equal(t, "raw-value", enc)
		assert.Equal(t, "raw-value", decryptLineCredential(ctx, nil, "raw-value"))
	})

	t.Run("empty value stays empty", func(t *testing.T) {
		enc, err := encryptLineCredential(cipher, "")
		assert.NoError(t, err)
		assert.Equal(t, "", enc)
		assert.Equal(t, "", decryptLineCredential(ctx, cipher, ""))
	})
}
