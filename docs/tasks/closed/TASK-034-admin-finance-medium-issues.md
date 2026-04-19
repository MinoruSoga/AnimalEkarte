# TASK-034: 管理・会計系 MEDIUM 問題 4件

## 優先度

MEDIUM

---

## 問題 1: clinic_handler の hasPermission が repository を直接呼んでいる

### ファイル
`backend/internal/handler/clinic_handler.go:62`

### 問題
```go
func (h *Handler) hasPermission(c *gin.Context, resource, action string) bool {
    // ...
    rules, err := h.repos.PermissionGroup.GetEffectivePermissionsByStaffID(c.Request.Context(), staffID)
```
ハンドラ共通ヘルパー `hasPermission` が `h.repos.PermissionGroup` を直接参照している。handler → repository の直接呼び出しは規約違反（TASK-031 の問題 1 と同根）。

### 修正案
`StaffService` または `PermissionGroupService` に `GetEffectivePermissions(ctx, staffID)` を追加し、`hasPermission` を service 経由で呼ぶ。

---

## 問題 2: refund_service の Create で billing.Status チェックが欠落

### ファイル
`backend/internal/service/refund_service.go:28-51`

### 問題
返金前に請求 (`billing`) の状態（`paid` か否か）を確認していない。`waiting` 状態の請求に対しても返金を作成できてしまう。過剰返金防止の残額チェックはあるが、状態マシン遷移の正当性チェックが欠落している。

### 修正案
```go
// 請求状態の確認（paid のみ返金可能）
if billing.Status != model.BillingStatusPaid {
    return nil, apperrors.WrapInvalidInput("支払済みの請求のみ返金できます")
}
```

---

## 問題 3: inventory_service の Create が *model.InventoryItem を受け取っている

### ファイル
`backend/internal/service/inventory_service.go`（`Create` メソッド）

### 問題
`InventoryService.Create(ctx, clinicID, item *model.InventoryItem)` のシグネチャが model を直接受け取っている。`handler/inventory_handler.go` が model を組み立てて渡すパターンになっており、handler 層でのビジネスロジック混入を招く。

### 修正案
`CreateInventoryInput` DTO を service に追加し、model 組み立ては service 内で完結させる（TASK-033 問題 2 の修正と連動）。

---

## 問題 4: accounting_service の小計整合性チェックが欠落

### ファイル
`backend/internal/service/accounting_service.go`（請求確定処理）

### 問題
請求確定時に `subtotal + tax_total = total_amount` の整合性チェックが行われていない。クライアント送信値をそのまま永続化するため、フロントエンドの計算誤りや改ざんを検出できない。

### 修正案
```go
// 確定前に整合性チェック
if input.Subtotal+input.TaxTotal != input.TotalAmount {
    return nil, apperrors.WrapInvalidInput("小計と税額の合計が請求合計と一致しません")
}
```
