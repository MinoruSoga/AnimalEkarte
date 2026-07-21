package lstep

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/infra/crypto"
)

// LINE 予約設定の line_channel_secret / line_access_token は保存時に AES-256-GCM で暗号化する（H-4）。
// lstep 連携（clinic_integrations）と同一の cipher（環境変数 INTEGRATION_ENCRYPTION_KEY）を再利用し、
// 独自の暗号実装は持たない。暗号文形式は crypto.AESGCMCipher と同一（base64(nonce + ciphertext)）。

// EncryptLineCredential は LINE 認証情報を暗号化する。
// cipher が nil（開発環境で INTEGRATION_ENCRYPTION_KEY 未設定）または value が空文字の場合はそのまま返す。
func EncryptLineCredential(cipher *crypto.AESGCMCipher, value string) (string, error) {
	if cipher == nil || value == "" {
		return value, nil
	}
	return cipher.Encrypt(value)
}

// DecryptLineCredential は暗号化済みの LINE 認証情報を復号する。
//
// レガシー平文レコード（暗号化導入前に保存された値）との後方互換のため、復号に失敗した場合は
// 平文とみなしてそのまま返し、slog.WarnContext で記録する。AES-256-GCM は認証付き暗号のため、
// 平文（非暗号文）を復号しようとすると base64 デコード失敗・認証タグ不一致のいずれかで必ずエラーになる。
//
// 一括 migration は行わない（bug.md H-4 の判断: STG は再入力運用でも可）。レガシー行は次回の設定
// 保存時に暗号化される（機会的再暗号化 — line_reservation_setting_service.go Save 参照）。
//
// cipher が nil または value が空文字の場合はそのまま返す。
func DecryptLineCredential(ctx context.Context, cipher *crypto.AESGCMCipher, value string) string {
	if cipher == nil || value == "" {
		return value
	}
	plaintext, err := cipher.Decrypt(value)
	if err != nil {
		// 値そのものはログに出さない（機密情報）。
		slog.WarnContext(ctx, "未暗号化のLINE認証情報レコードを検出しました（次回の設定保存時に暗号化されます）")
		return value
	}
	return plaintext
}
