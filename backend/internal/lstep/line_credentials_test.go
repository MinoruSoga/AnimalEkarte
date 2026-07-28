package lstep

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/infra/crypto"
)

// TestLineCredentialEncryptDecrypt は LINE 認証情報の暗号化/復号ヘルパーを検証する（H-4）。
// testIntegrationKeyHex は be9_2c_r1_mock_carriers_test.go（R⑥でreservationへ移動した旧line_reservation_setting_service_test.go由来のcarrier） で定義。
func TestLineCredentialEncryptDecrypt(t *testing.T) {
	ctx := context.Background()
	cipher, err := crypto.NewAESGCMCipher(testIntegrationKeyHex)
	if err != nil {
		t.Fatalf("failed to build test cipher: %v", err)
	}

	t.Run("round-trip: encrypted value decrypts back to plaintext", func(t *testing.T) {
		enc, err := EncryptLineCredential(cipher, "channel-secret-123")
		assert.NoError(t, err)
		assert.NotEqual(t, "channel-secret-123", enc)

		got := DecryptLineCredential(ctx, cipher, enc)
		assert.Equal(t, "channel-secret-123", got)
	})

	t.Run("legacy plaintext row falls back to the raw value", func(t *testing.T) {
		// 暗号化導入前に保存された平文（AES-GCM 復号に失敗する）
		// 実クレデンシャルを使わず、明らかに合成の 32-hex ダミーを用いる
		const legacyPlain = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		got := DecryptLineCredential(ctx, cipher, legacyPlain)
		assert.Equal(t, legacyPlain, got)
	})

	t.Run("nil cipher passes the value through unchanged", func(t *testing.T) {
		enc, err := EncryptLineCredential(nil, "raw-value")
		assert.NoError(t, err)
		assert.Equal(t, "raw-value", enc)
		assert.Equal(t, "raw-value", DecryptLineCredential(ctx, nil, "raw-value"))
	})

	t.Run("empty value stays empty", func(t *testing.T) {
		enc, err := EncryptLineCredential(cipher, "")
		assert.NoError(t, err)
		assert.Equal(t, "", enc)
		assert.Equal(t, "", DecryptLineCredential(ctx, cipher, ""))
	})
}

// testIntegrationKeyHex は 32 バイト（AES-256）のダミー暗号鍵（service 側同名 const の複製）。
const testIntegrationKeyHex = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
