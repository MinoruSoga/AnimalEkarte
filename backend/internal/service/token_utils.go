package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// generateSecureToken は暗号学的に安全な32バイトのランダムトークンを生成する
func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", apperrors.Wrap(err, "generate secure token")
	}
	return hex.EncodeToString(b), nil
}

// sha256Hex はトークン文字列のSHA-256ハッシュを16進数文字列で返す
func sha256Hex(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
