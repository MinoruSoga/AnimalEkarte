// Package crypto はAES-256-GCM暗号化/復号化を提供する。
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

// AESGCMCipher はAES-256-GCMによる対称暗号化を提供する。
// キーは環境変数 INTEGRATION_ENCRYPTION_KEY（32バイトのhex文字列）から取得する。
type AESGCMCipher struct {
	key []byte
}

// NewAESGCMCipher は hex エンコードされた 32 バイトキーから AESGCMCipher を生成する。
// keyHex は 64 文字の hex 文字列（32 バイト = AES-256）でなければならない。
func NewAESGCMCipher(keyHex string) (*AESGCMCipher, error) {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key format: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes (64 hex chars), got %d bytes", len(key))
	}
	return &AESGCMCipher{key: key}, nil
}

// Encrypt は平文を AES-256-GCM で暗号化し、base64 エンコードされた文字列（nonce + ciphertext）を返す。
// APIキー等の機密情報のログ出力を防ぐため、引数/戻り値はログに含めないこと。
func (c *AESGCMCipher) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt は Encrypt で生成した base64 文字列を復号して平文を返す。
func (c *AESGCMCipher) Decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	return string(plaintext), nil
}

// MaskValue は機密値をマスクして末尾4文字のみ返す。
// 4文字未満の場合は全てマスクする。
func MaskValue(v string) string {
	if len(v) <= 4 {
		return "****"
	}
	return "****" + v[len(v)-4:]
}
