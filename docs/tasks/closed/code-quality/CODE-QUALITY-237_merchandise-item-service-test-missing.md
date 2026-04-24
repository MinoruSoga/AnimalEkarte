# CODE-QUALITY-237: merchandise_item_service テスト — Create/Update テスト欠落

## 概要

`backend/internal/service/merchandise_item_service_test.go` に
Delete テストのみ存在し、Create・Update のテストが完全に欠落している。

---

## 現状

```
merchandise_item_service_test.go の実装:
- TestMerchandiseItemService_Delete (5パターン) ✅
- TestMerchandiseItemService_Create             ❌ 欠落
- TestMerchandiseItemService_Update             ❌ 欠落
```

---

## 欠落している検証内容

### Create テスト
- 正常作成
- 名前バリデーション（空文字 → エラー）
- 金額バリデーション（price < 0 → エラー）
- 重複名 → 409 Conflict

### Update テスト
- nil input → `apperrors.ErrInvalidInput`
- 存在しない ID → 404 Not Found
- 正常更新
- 金額バリデーション（price < 0 → エラー）
- ErrMsgAtLeastOneField チェック（全フィールド nil）

---

## 参考（merchandise_item_service.go の検証ロジック）

```go
// Create
func (s *merchandiseItemService) Create(...) {
    if err := validateRequiredName(input.Name); err != nil { ... }
    if input.Price != nil && *input.Price < 0 {
        return nil, apperrors.WrapInvalidInput("金額は0以上を入力してください")
    }
    // ...
}

// Update
func (s *merchandiseItemService) Update(...) {
    if input == nil { return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil) }
    if _, err := s.repo.FindByID(...); err != nil { ... }
    if input.Price != nil && *input.Price < 0 { ... }
    if len(fields) == 0 { return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField) }
    // ...
}
```

---

## 優先度

MEDIUM — 金額バリデーション・nil チェック・FK 競合チェックがテストで検証されておらず、
リグレッションリスクがある。billing_item_service と合わせてテストを整備すべき。
