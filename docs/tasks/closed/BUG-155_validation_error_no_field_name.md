# BUG-155: バリデーションエラーメッセージにフィールド名がなく、どの項目が不足か判別できない

## 概要
BUG-135 の修正でバリデーションエラーメッセージが汎化されたが、過度に汎化されてしまい
「必須項目が入力されていません」としか返さなくなった。
どのフィールドが不足しているか開発者にもユーザーにもわからない。

複数フィールドが不足している場合、同じメッセージがセミコロン区切りで繰り返される:
`必須項目が入力されていません; 必須項目が入力されていません; 必須項目が入力されていません`

## 再現手順
```bash
curl -X POST /api/v1/hospitalizations -H 'Content-Type: application/json' -d '{}'
# → {"error":"必須項目が入力されていません; 必須項目が入力されていません; ..."}

# 5回繰り返されるが、どの5フィールドが不足しているか不明
```

## 期待する動作

### 案A: フィールドの日本語ラベルを返す
```json
{"error": "ペット、飼主、ケージ、入院日、ステータスは必須です"}
```

### 案B: フィールドキーを返す（API 利用者向け）
```json
{
  "error": "入力値が正しくありません",
  "fields": {
    "pet_id": "必須です",
    "owner_id": "必須です",
    "cage_id": "必須です",
    "admission_date": "必須です",
    "status": "必須です"
  }
}
```

### 案C: 汎化メッセージ + フィールド数
```json
{"error": "5件の必須項目が入力されていません"}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/error-handling.md`
> Error Response Format: `{"error": "Human readable message"}`

現状は human readable ではない（どのフィールドか不明）。

### `.claude/rules/api.md`
> "Use meaningful error messages"

## 優先度
**Medium** — 開発者・フロントエンドがどのフィールドを修正すべきか判別できない。
BUG-135 修正の副作用。

## 関連チケット
- BUG-135（修正済み）: Go 構造体フィールド名漏洩の修正が過度に汎化

## 関連ファイル
- `backend/internal/handler/response.go` — parseBindError
