# BUG-345: バックエンド Go コード規約準拠監査 第8回（2026-04-14）

## 概要

`backend/` 全ドメインを対象にコード規約準拠監査を実施した。前回（第7回: BUG-338〜340）以降の変更を含む最新コードを検証。今回は **CRITICAL/HIGH ゼロ**。残存するのは **ファイル行数超過 1件（Low）** のみ。

## 監査スコープ

| カテゴリ | チェック内容 | 結果 |
|---------|-------------|------|
| Context 伝播 | 全 DB 呼び出しで `WithContext(ctx)` 使用 | ✅ 違反ゼロ |
| エラーハンドリング | `apperrors.FromGORM` / `Wrap` 使用 | ✅ 違反ゼロ |
| Handler レスポンス | `RespondError(c, err)` 使用 / `c.JSON(4xx/5xx)` 直接使用なし | ✅ 1件は意図的例外（注記あり） |
| slog ロギング | handler/repository 層に slog 記述なし（許容例外のみ） | ✅ 違反ゼロ |
| マスタ削除 FK チェック | 全マスタ Delete に `WrapConflict` 依存チェックあり | ✅ 違反ゼロ |
| グローバル変数 | ミュータブルなグローバル状態なし | ✅ 違反ゼロ（`lineVerifyURL` はテスト用意図的) |
| context.Background() | 通知サービスのゴルーチン以外で不使用 | ✅ 違反ゼロ |
| エラー握りつぶし | `_ = err` なし | ✅ 違反ゼロ |
| panic | テスト外での `panic()` なし | ✅ 違反ゼロ |
| ファイル行数 | 1ファイル < 500行 | ⚠️ 3ファイル超過 → **BUG-346** |

## 意図的例外（違反ではない）

### `liff_handler.go:207` — `c.JSON(http.StatusConflict, ...)`
```go
// NOTE: Intentional direct response — ReservationLimitError は RespondError で処理できないカスタムペイロード（redirect_step）を含む
c.JSON(http.StatusConflict, gin.H{
    "error":         limErr.Error(),
    "code":          limErr.Code,
    "redirect_step": limErr.RedirectStep,
})
```
`ReservationLimitError` は `redirect_step` フィールドを含むカスタムペイロードが必要。`RespondError` の標準フォーマットでは表現できないため、意図的な直接使用。適切なコメントが付いている。

### `appointment_notification_service.go:53,100` — `context.Background()` 使用
通知は HTTP リクエスト完了後も継続するため、独立した background context を使用。`//nolint:contextcheck,gosec` コメントで意図を明示済み。

### `middleware/liff_auth.go:137` — `var lineVerifyURL = "..."`
テスト用 URL オーバーライドのためのパッケージレベル変数。同パッケージの `liff_auth_test.go` で使用。Go の標準的テストパターン。

## 派生チケット

| BUG | 内容 | 優先度 |
|-----|------|--------|
| [BUG-346](BUG-346_file-size-over-500-lines.md) | ファイル行数超過（auth_handler.go 617行 / staff_handler.go 609行 / liff_service.go 555行） | Low |

## 監査完了判定

**第8回監査: PASS（Low 1件のみ）**

前回監査からの改善確認:
- `BUG-339` (JOIN 先 `deleted_at IS NULL` 欠落): 実コードで全修正確認済み
- `BUG-340` (`clinicScope` 未使用): 実コードで修正確認済み
- 全リポジトリが `clinicScope` または table-qualified `clinic_id` を正しく使用
