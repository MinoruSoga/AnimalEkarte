# BUG-145: 既存メールアドレスで重複スタッフ・アカウントが作成可能

## 概要
`POST /api/v1/masters/staffs` で既に使用されているメールアドレスを指定しても 201 で作成される。
同一メールの accounts レコードが作成されるか、account なしのスタッフが作成される。
いずれの場合もデータ整合性の問題。

## 脆弱性分類
- **CWE-20**: Improper Input Validation
- **影響**: 同一メールの複数アカウント → ログイン時の混乱、アカウント管理の不整合

## 再現手順
```bash
# 既存メール admin@example.com で新スタッフ作成
curl -X POST /api/v1/masters/staffs \
  -H 'Content-Type: application/json' \
  -d '{"name": "dup_email_test", "email": "admin@example.com"}'
# → 201 Created ❌
```

## 期待する動作
- 409 Conflict: `このメールアドレスは既に使用されています`

## 修正方針

### Service 層で重複チェック
```go
func (s *StaffService) Create(ctx context.Context, input CreateInput) (*model.Staff, error) {
    if input.Email != "" {
        existing, _ := s.accountRepo.FindByEmail(ctx, input.Email)
        if existing != nil {
            return nil, apperrors.WrapConflict("このメールアドレスは既に使用されています")
        }
    }
    // ...
}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/database-design.md`
> **UNIQUE 制約: 論理削除対応（部分インデックス）**

`accounts.email` に UNIQUE 制約が設定されているか確認が必要。
設定されていれば DB レベルで 500 になるはず（→ BUG-138 で 400 に変換すべき）。

### `.claude/CLAUDE.md` — エラー処理の統一 (MANDATORY)
> マスタ削除時は必ず依存レコードの存在をチェックし、参照がある場合は `apperrors.WrapConflict(...)` で 409 を返す

作成時の重複チェックも同じパターン。

## 優先度
**Medium** — データ整合性の問題。同一メールの複数スタッフが運用上の混乱を引き起こす。

## 関連ファイル
- `backend/internal/service/staff_service.go` — Create
- `backend/internal/handler/staff_handler.go` — CreateStaff
- `backend/migrations/001_init.sql` — accounts テーブル UNIQUE 制約
