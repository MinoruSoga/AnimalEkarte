# BUG-135: バリデーションエラーに Go 構造体フィールド名が漏洩する

## 概要
必須フィールド欠落時のバリデーションエラーメッセージに Go 構造体の json タグ名がそのまま返される。
BUG-129 で JSON パースエラーと not found メッセージは修正されたが、
`binding:"required"` 由来のバリデーションエラーには修正が適用されていない。

## 脆弱性分類
- **CWE-209**: Generation of Error Message Containing Sensitive Information
- **影響**: API の内部フィールド名が推測可能。攻撃の足がかりになる。

## 再現手順

### 1. Owner 作成 — name 欠落
```bash
curl -X POST /api/v1/owners \
  -H 'Content-Type: application/json' \
  -d '{"phone": "000"}'
```
**レスポンス**:
```json
{"error": "owner_name is required"}
```
**漏洩**: `owner_name` — Go 構造体の json タグ名

### 2. Reservation 作成 — 必須フィールド欠落
```bash
curl -X POST /api/v1/reservations \
  -H 'Content-Type: application/json' \
  -d '{}'
```
**レスポンス**:
```json
{"error": "start_time is required; end_time is required; service_type_i_d is required"}
```
**漏洩**:
- `start_time`, `end_time` — フィールド名
- `service_type_i_d` — **Go 構造体の `ServiceTypeID` の snake_case 変換ミス**（`service_type_id` であるべき）

### 3. Whitespace-only name
```bash
curl -X POST /api/v1/owners \
  -H 'Content-Type: application/json' \
  -d '{"name": "   ", "phone": "000"}'
```
**レスポンス**:
```json
{"error": "owner_name is required"}
```

## 期待するレスポンス

```json
{"error": "入力値が正しくありません", "fields": {"name": "必須項目です"}}
```

または日本語のみ:
```json
{"error": "名前は必須です"}
```

## 現状コード

### `backend/internal/handler/response.go` — parseBindError

BUG-129 で `json.SyntaxError` と `json.UnmarshalTypeError` は修正されたが、
`validator.ValidationErrors`（gin の `binding:"required"` 由来）のサニタイズが不完全。

```go
func parseBindError(err error) string {
    var ve validator.ValidationErrors
    if errors.As(err, &ve) {
        // ❌ json タグ名がそのまま返される
        msgs := make([]string, 0, len(ve))
        for _, fe := range ve {
            msgs = append(msgs, fmt.Sprintf("%s is required", fe.Field()))
            // fe.Field() → "owner_name", "service_type_i_d" etc.
        }
        return strings.Join(msgs, "; ")
    }
}
```

## 修正方針

バリデーションエラーのフィールド名を汎化するか、ユーザーフレンドリーなラベルに変換:

```go
func parseBindError(err error) string {
    var ve validator.ValidationErrors
    if errors.As(err, &ve) {
        // フィールド名を返さず、汎化メッセージ
        return "入力値が正しくありません"
    }
}
```

または、フィールド名 → 日本語ラベルのマッピングを用意:

```go
var fieldLabels = map[string]string{
    "Name":          "名前",
    "Phone":         "電話番号",
    "StartTime":     "開始時間",
    "EndTime":       "終了時間",
    "ServiceTypeID": "診療サービス",
}
```

## 付随する問題: `service_type_i_d` の snake_case 変換ミス

`ServiceTypeID` が `service_type_id` ではなく `service_type_i_d` に変換されている。
これは gin の validator が Go フィールド名を自動で snake_case に変換する際の問題。
json タグに `json:"service_type_id"` が設定されていても、validator はフィールド名（PascalCase）を使う。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — エラー処理の統一 (MANDATORY)
> `RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))` （全31ハンドラ統一済み）

parseBindError の出力をサニタイズすれば全ハンドラに自動適用される。

### `.claude/rules/security.md` — Logging
> "Never log sensitive data"

内部フィールド名はログに記録し、ユーザーには汎化メッセージを返す。

## 優先度
**Low** — BUG-129 の残対応。フィールド名の漏洩は直接的な攻撃にはつながりにくい。

## 関連チケット
- BUG-129（修正済み）: エラーメッセージ Go 内部情報漏洩

## 関連ファイル
- `backend/internal/handler/response.go` — parseBindError
