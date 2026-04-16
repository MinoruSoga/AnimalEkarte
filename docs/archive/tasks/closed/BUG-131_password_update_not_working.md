# BUG-131: スタッフのパスワード更新が機能していない

## 概要
`PATCH /api/v1/masters/staffs/:id` で `password` フィールドを送信しても
`at least one field must be provided` エラーが返る。パスワード更新が全く機能していない。

## 脆弱性分類
- **機能バグ**（セキュリティ関連）
- **影響**: 管理者がスタッフのパスワードを変更できない

## 再現手順
```bash
# 7文字パスワード
curl -X PATCH http://localhost:8080/api/v1/masters/staffs/9 \
  -H 'Content-Type: application/json' \
  -b 'auth_token=<JWT>' \
  -d '{"password": "1234567"}'
# → 400 {"error": "at least one field must be provided"}

# 8文字パスワード
curl -X PATCH http://localhost:8080/api/v1/masters/staffs/9 \
  -H 'Content-Type: application/json' \
  -b 'auth_token=<JWT>' \
  -d '{"password": "12345678"}'
# → 400（同じエラー）
```

## ブラウザテスト結果

| テスト | パスワード | 期待 | 実際 |
|--------|----------|------|------|
| 7文字 | `1234567` | 400 (too short) | 400 `at least one field must be provided` |
| 8文字 | `12345678` | 200 (success) | **400** |

## 原因分析

### `backend/internal/handler/staff_request.go`

```go
type updateStaffRequest struct {
    Name          *string `json:"name"`
    LicenseNumber *string `json:"license_number"`
    OccupationID  *uint64 `json:"occupation_id"`
    SortOrder     *int    `json:"sort_order"`
    IsActive      *bool   `json:"is_active"`
    Password      *string `json:"password"`
}
```

`Password` フィールドは定義されている。問題は service 層の `buildUpdateFields` か
handler の `UpdateStaff` 内の分岐ロジックにある。

### 推定原因
`password` は `buildStaffUpdateFields()` で `map[string]any` に含まれない可能性がある。
パスワードは DB カラムではなく、`accounts.password_hash` を bcrypt で更新する特別処理が必要。
`buildStaffUpdateFields()` が `password` フィールドを無視し、他のフィールドも nil なので
「少なくとも1フィールド必要」エラーになる。

## 修正方針

`UpdateStaff` ハンドラで `password` を `buildStaffUpdateFields` とは別に処理する:

```go
func (h *Handler) UpdateStaff(c *gin.Context) {
    // ... req バインド

    // password は別処理
    if req.Password != nil && *req.Password != "" {
        // password のみ送信された場合も「フィールドなし」エラーにしない
    }
    
    fields := buildStaffUpdateFields(req)
    hasPasswordUpdate := req.Password != nil && *req.Password != ""
    
    if len(fields) == 0 && !hasPasswordUpdate {
        RespondError(c, apperrors.WrapInvalidInput("at least one field must be provided"))
        return
    }
    
    // ... fields の更新
    // ... password の更新（bcrypt）
}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/security.md` — Authentication
> "Implement proper password hashing (bcrypt/argon2)"

パスワード更新自体が機能していないため、この規約を満たせていない。

### `.claude/CLAUDE.md` — エラー処理の統一
`at least one field must be provided` は password のみの更新を想定していないロジックの問題。

## 優先度
**High** — 管理者がスタッフのパスワードをリセットできない。運用上の問題。

## 関連ファイル
- `backend/internal/handler/staff_handler.go` — UpdateStaff ハンドラ
- `backend/internal/handler/staff_request.go` — updateStaffRequest 定義
- `backend/internal/service/staff_service.go` — buildStaffUpdateFields
