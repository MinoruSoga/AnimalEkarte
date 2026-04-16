# BUG-129: API エラーレスポンスに Go 内部情報が漏洩している

## 概要
API のエラーレスポンスに Go の内部実装詳細（構造体名、フィールド名、パーサーエラー）が
そのまま返されている。攻撃者に API の内部構造を推測させる情報漏洩。

## 脆弱性分類
- **CWE-209**: Generation of Error Message Containing Sensitive Information
- **OWASP A05:2021**: Security Misconfiguration
- **影響**: API 内部構造の推測が可能。直接的なデータ漏洩ではないが、攻撃の足がかりになる。

## 再現手順と漏洩内容

### 1. JSON パースエラー — Go パーサーエラーメッセージがそのまま
```bash
curl -X POST /api/v1/owners -H 'Content-Type: application/json' -d '{invalid}'
```
**レスポンス**:
```json
{"error": "invalid character 'i' looking for beginning of object key string"}
```
**漏洩情報**: Go `encoding/json` パーサーの内部エラーメッセージ

### 2. 型不一致エラー — Go 構造体名とフィールド名が漏洩
```bash
curl -X POST /api/v1/owners -H 'Content-Type: application/json' \
  -d '{"name": 123, "phone": true}'
```
**レスポンス**:
```json
{"error": "json: cannot unmarshal bool into Go struct field createOwnerRequest.phone of type string"}
```
**漏洩情報**: 
- Go 構造体名: `createOwnerRequest`
- フィールド名: `phone`
- フィールド型: `string`

### 3. リソース未存在 — リソース名と ID が漏洩
```bash
curl /api/v1/owners/999999
```
**レスポンス**:
```json
{"error": "owner with id 999999 not found"}
```
**漏洩情報**: リソース名 `owner`、ID `999999`

```bash
curl /api/v1/medical-records/999999
```
**レスポンス**:
```json
{"error": "medical_record with id 999999 not found"}
```
**漏洩情報**: テーブル名推測可能 `medical_record`

```bash
curl /api/v1/masters/staffs/999999
```
**レスポンス**:
```json
{"error": "staff with id 999999 not found"}
```

## 期待する動作

### JSON パースエラー
```json
{"error": "リクエストの形式が正しくありません"}
```

### 型不一致エラー
```json
{"error": "入力値の形式が正しくありません", "fields": {"phone": "文字列で指定してください"}}
```

### リソース未存在
```json
{"error": "指定されたリソースが見つかりません"}
```
またはシンプルに:
```json
{"error": "not found"}
```

## 現状コード

### `backend/internal/handler/response.go` — parseBindError

```go
// ShouldBindJSON エラーを RespondError 経由で返している
// ただし Go の内部エラーメッセージがそのまま error フィールドに入る
RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
```

### `backend/internal/errors/errors.go` — FromGORM

```go
// FromGORM は GORM エラーを変換するが、リソース名と ID をエラーメッセージに含める
func FromGORM(err error, resource string, id string) error {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return fmt.Errorf("%s with id %s not found: %w", resource, id, ErrNotFound)
    }
    // ...
}
```

## 修正方針

### 1. JSON バインドエラーのサニタイズ

`parseBindError()` の返り値を Go 内部メッセージからユーザーフレンドリーなメッセージに変換:

```go
func parseBindError(err error) string {
    var unmarshalErr *json.UnmarshalTypeError
    if errors.As(err, &unmarshalErr) {
        return fmt.Sprintf("%s: 正しい形式で入力してください", unmarshalErr.Field)
    }
    var syntaxErr *json.SyntaxError
    if errors.As(err, &syntaxErr) {
        return "リクエストの形式が正しくありません"
    }
    // Go 内部メッセージは返さない
    return "入力値が正しくありません"
}
```

### 2. FromGORM の not found メッセージを汎化

```go
func FromGORM(err error, resource string, id string) error {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        // 内部ログにはリソース名と ID を記録
        slog.Info("resource not found", "resource", resource, "id", id)
        // ユーザーには汎化メッセージを返す
        return fmt.Errorf("not found: %w", ErrNotFound)
    }
}
```

### 3. RespondError の最終フィルタ

```go
func RespondError(c *gin.Context, err error) {
    // ... status code 判定
    
    // ユーザー向けメッセージにフォールバック
    userMessage := getUserFriendlyMessage(err, code)
    c.JSON(code, gin.H{"error": userMessage})
}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/security.md` — Logging
> "Never log sensitive data (passwords, tokens)"

エラーメッセージも同様。内部実装の詳細はログに記録し、ユーザーには汎化メッセージを返す。

### `.claude/CLAUDE.md` — エラー処理の統一 (MANDATORY)
> **Handler: `RespondError(c, err)` で統一レスポンス**

RespondError 内でエラーメッセージのサニタイズを行えば、全エンドポイントに自動適用される。

### `.claude/rules/error-handling.md`
> Error Response Format: `{"error": "Human readable message"}`

"Human readable" は「ユーザーが理解できるメッセージ」であり、Go 内部エラーではない。

## 優先度
**Medium** — 直接的なデータ漏洩ではないが、攻撃者に内部構造を推測させる情報を提供。
本番デプロイ前に修正推奨。

## 関連ファイル
- `backend/internal/handler/response.go` — RespondError, parseBindError
- `backend/internal/errors/errors.go` — FromGORM, エラーメッセージ生成
