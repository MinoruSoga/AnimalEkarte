# TASK-213: payment_method_master_handler.go — DELETE ルートに誤った権限（edit）が設定されている

## 優先度
High

## 対象ファイル
`backend/internal/handler/payment_method_master_handler.go`

## 問題概要
Delete エンドポイントのルート定義に `can_edit` 権限が設定されており、`can_delete` が正しい。
`edit` 権限のみ持つスタッフが DELETE 操作を実行できてしまう認可バグ。

## 現状コード（行145付近）

```go
// 現状（NG）
pm.DELETE("/:id",
    h.RequirePermission(string(model.ResourcePaymentMethod), "edit"),
    h.DeletePaymentMethod,
)
```

## あるべき姿

```go
// あるべき姿
pm.DELETE("/:id",
    h.RequirePermission(string(model.ResourcePaymentMethod), "delete"),
    h.DeletePaymentMethod,
)
```

## 影響範囲
- 編集権限のみ持つロールが支払方法を削除できてしまう
- 最小権限の原則に違反

## 完了条件
- [ ] DELETE ルートの権限を `"edit"` → `"delete"` に修正
- [ ] 他の DELETE ルートに同様の誤設定がないか全ハンドラを確認
- [ ] `go test ./backend/internal/...` がパス
