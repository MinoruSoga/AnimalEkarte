---
title: Handler P5/P6 Routes RequirePermission Middleware 欠落
issue: '#474'
priority: CRITICAL
status: open
area: handler
pattern: P5/P6
---

## 概要

ハンドラーレイヤーの POST/PUT/PATCH/DELETE ルート定義で、権限検証ミドルウェア `RequirePermission` が欠落しているケースが 2 件検出されました。

### パターン
- **P5 違反**：POST/PUT/PATCH ルートに `RequirePermission("edit")` がない
- **P6 違反**：DELETE ルートに `RequirePermission("delete")` がない（"edit" では不十分）

### 違反ファイル一覧

| ファイル | 行番号 | ルート | メソッド | 問題 | 修正権限 |
|---------|--------|--------|---------|------|---------|
| payment_method_master_handler.go | 45 | POST /clinics/:clinicID/payment-methods | Create | RequirePermission がない | "edit" |
| clinic_holiday_handler.go | 62 | DELETE /clinics/:clinicID/holidays/:id | Delete | RequirePermission("edit") ❌ 必須は "delete" | "delete" |

## 修正方法

### P5 違反（POST/PUT/PATCH）
```go
// payment_method_master_handler.go (L45 付近)
router.POST("/clinics/:clinicID/payment-methods",
    middleware.RequirePermission("edit"),  // ← 追加
    h.Create)
```

### P6 違反（DELETE）
```go
// clinic_holiday_handler.go (L62 付近)
router.DELETE("/clinics/:clinicID/holidays/:id",
    middleware.RequirePermission("delete"),  // ← "edit" ではなく "delete" に変更
    h.Delete)
```

## テスト

修正後、以下の確認を実施：
- [ ] "edit" 権限ユーザーが POST/PUT/PATCH を実行できること
- [ ] "delete" 権限がないユーザーが DELETE を実行時に 403 を返すこと
- [ ] 権限なしユーザーが POST/PUT/PATCH/DELETE 実行時に 403 を返すこと
- [ ] インテグレーションテストが全件パス

## 参考

- Pattern: P5 (RequirePermission "edit" for POST/PUT/PATCH)
- Pattern: P6 (RequirePermission "delete" for DELETE)
- 関連: Routes レイヤー権限検証 MUST
