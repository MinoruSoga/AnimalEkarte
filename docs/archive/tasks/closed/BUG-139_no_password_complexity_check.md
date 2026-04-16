# BUG-139: パスワード複雑性チェックがない（長さのみ）

## 概要
パスワード更新時のバリデーションが「8文字以上」のみ。
`11111111` や `aaaaaaaa` のような脆弱なパスワードが設定可能。

## 脆弱性分類
- **CWE-521**: Weak Password Requirements
- **OWASP A07:2021**: Identification and Authentication Failures
- **影響**: ブルートフォース攻撃（BUG-130 で緩和済み）や辞書攻撃に対して脆弱

## 再現手順
```bash
curl -X PATCH /api/v1/masters/staffs/9 \
  -H 'Content-Type: application/json' \
  -d '{"password": "11111111"}'
# → 200 OK（設定成功）

curl -X PATCH /api/v1/masters/staffs/9 \
  -H 'Content-Type: application/json' \
  -d '{"password": "aaaaaaaa"}'
# → 200 OK（設定成功）
```

## ブラウザテスト結果
| パスワード | 期待 | 実際 |
|-----------|------|------|
| `1234567` (7文字) | 400 | 400 ✅ |
| `12345678` (8文字) | 200 | 200 ✅ |
| `11111111` (全数字) | 400 | **200** ⚠️ |
| `aaaaaaaa` (全同一文字) | 400 | **200** ⚠️ |

## 期待する動作
- 8文字以上
- 英字（大文字 or 小文字）+ 数字を含む
- 同一文字の連続3回以上を禁止
- 一般的な辞書パスワード（password, 12345678 等）を禁止

## 修正方針

```go
func validatePassword(password string) error {
    if len(password) < 8 {
        return fmt.Errorf("パスワードは8文字以上で入力してください")
    }
    hasLetter := false
    hasDigit := false
    for _, c := range password {
        if unicode.IsLetter(c) { hasLetter = true }
        if unicode.IsDigit(c) { hasDigit = true }
    }
    if !hasLetter || !hasDigit {
        return fmt.Errorf("パスワードは英字と数字を含めてください")
    }
    return nil
}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/security.md` — Authentication
> "Implement proper password hashing (bcrypt/argon2)"

ハッシュ化は実装済みだが、入力時の複雑性チェックが不足。

## 優先度
**Low** — BUG-130 のレートリミットで緩和済み。本番環境ではパスワードポリシーの強化を推奨。

## 関連チケット
- BUG-130（修正済み）: レートリミット
- BUG-131（修正済み）: パスワード更新機能

## 関連ファイル
- `backend/internal/handler/staff_handler.go` — パスワード更新ロジック
