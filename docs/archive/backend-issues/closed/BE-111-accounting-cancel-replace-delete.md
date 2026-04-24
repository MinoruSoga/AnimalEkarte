# BE-111: 精算済会計の論理削除 API（DELETE 廃止 / POST cancel 新設）

**Status**: Closed (2026-04-14, commit 8fcd1382)
**Priority**: High
**Affects**: `accounting_service.go`, `accounting_handler.go`, `handler.go`
**Date Created**: 2026-04-14
**Related**: BUG-371, FE-249

## Summary

ハード削除の `DELETE /accountings/:id` を撤去し、論理削除の `POST /accountings/:id/cancel` に置き換える。Update API は status 制限を持たないため変更不要。

## 現状のコード

### 既存ハード削除 ルート
```go
// backend/internal/handler/handler.go:180
accountings.DELETE("/:id", h.RequirePermission(string(model.ResourceAccounting), "delete"), h.DeleteAccounting)
```

### 既存ハンドラ
```go
// backend/internal/handler/accounting_handler.go:192-207
func (h *Handler) DeleteAccounting(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok { return }
    id, ok := parseIDParam(c, "id")
    if !ok { return }
    if err := h.svc.Accounting.Delete(c.Request.Context(), clinicID, id); err != nil {
        RespondError(c, err)
        return
    }
    c.Status(http.StatusNoContent)
}
```

### 既存サービス（ハード削除 + FK 依存チェック）
```go
// backend/internal/service/accounting_service.go:257-276
func (s *accountingService) Delete(ctx context.Context, clinicID, id uint64) error {
    itemCount, err := s.repo.CountItemsByBillingID(ctx, clinicID, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to check billing item dependencies")
    }
    if itemCount > 0 {
        return apperrors.WrapConflict("請求明細が紐付いているため削除できません。先に請求明細を削除してください")
    }
    if err := s.repo.Delete(ctx, clinicID, id); err != nil {
        return apperrors.Wrap(err, "failed to delete accounting")
    }
    slog.InfoContext(ctx, "billing deleted", ...)
    return nil
}
```

### Update API（status 制限なし — 変更不要）
```go
// backend/internal/service/accounting_service.go:165-202
// status, completed_at を含む全フィールドを更新可能。
// 完了後の修正で「completed のまま」にする場合、FE は status を渡さなければよい。
```

## 必要な変更

### 1. DB マイグレーション
**なし**

### 2. Model 変更
**なし**

### 3. Service 変更

```go
// backend/internal/service/accounting_service.go

// インターフェース変更
type AccountingService interface {
    List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error)
    GetByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
    Create(ctx context.Context, input *CreateAccountingInput) (*model.Billing, error)
    Update(ctx context.Context, input *UpdateAccountingInput) (*model.Billing, error)
    Cancel(ctx context.Context, clinicID, id uint64) (*model.Billing, error)  // ★ Delete を Cancel に置換
    // ListUnpaidByOwner / ListUnpaidByBilling は BE-110 で追加予定
}

// Cancel 実装（ハード削除 → 論理削除へ変更）
func (s *accountingService) Cancel(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
    // 既存レコード存在チェック（マルチテナント絞り込み込み）
    existing, err := s.repo.FindByID(ctx, clinicID, id)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to find accounting for cancel")
    }

    // 既に cancelled なら 409
    if existing.Status == model.BillingStatusCancelled {
        return nil, apperrors.WrapConflict("この会計は既にキャンセル済みです")
    }

    // status のみ更新（completed_at, total_amount 等は保持して監査性確保）
    fields := map[string]any{
        "status": model.BillingStatusCancelled,
    }
    updated, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to cancel accounting")
    }

    slog.InfoContext(ctx, "accounting cancelled",
        slog.Uint64("billing_id", id),
        slog.Uint64("clinic_id", clinicID),
        slog.String("previous_status", string(existing.Status)),
    )
    return updated, nil
}

// 旧 Delete メソッドは削除する（インターフェース・実装両方）。
// CountItemsByBillingID は他から呼ばれていなければ削除（要 grep 確認）。
```

### 4. Handler 変更

```go
// backend/internal/handler/accounting_handler.go

// 旧 DeleteAccounting を CancelAccounting にリネーム
func (h *Handler) CancelAccounting(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    id, ok := parseIDParam(c, "id")
    if !ok {
        return
    }
    cancelled, err := h.svc.Accounting.Cancel(c.Request.Context(), clinicID, id)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, toAccountingResponse(cancelled))
}
```

### 5. ルート変更

```go
// backend/internal/handler/handler.go:180

// Before:
// accountings.DELETE("/:id", h.RequirePermission(string(model.ResourceAccounting), "delete"), h.DeleteAccounting)

// After:
accountings.POST("/:id/cancel", h.RequirePermission(string(model.ResourceAccounting), "delete"), h.CancelAccounting)
```

## API レスポンス形式

```json
// POST /api/clinics/1/accountings/100/cancel
// 成功: 200 OK
{
  "id": 100,
  "clinic_id": 1,
  "status": "cancelled",
  "scheduled_date": "2026-04-10T00:00:00+09:00",
  "completed_at": "2026-04-10T15:30:00+09:00",
  "total_amount": 5000,
  ...
}

// 既にキャンセル済: 409 Conflict
{
  "code": "conflict",
  "message": "この会計は既にキャンセル済みです",
  "timestamp": "..."
}

// 権限なし: 403 Forbidden
// 存在しない: 404 Not Found
```

## フロントエンド影響

- `make codegen` 不要（モデル変更なし）
- FE-249 で `DELETE /accountings/:id` の代わりに `POST /accountings/:id/cancel` を呼ぶ実装が必要
- 既存 `AccountingList` 画面の `cancelled` フィルタは既存（`AccountingList.tsx:78`）

## 完了条件

- [ ] DB マイグレーションなし
- [ ] `AccountingService.Delete` を `Cancel` にリネーム + 実装変更
- [ ] `DeleteAccounting` ハンドラを `CancelAccounting` にリネーム
- [ ] ルート: `DELETE /:id` を撤去、`POST /:id/cancel` を登録
- [ ] テストケース追加（テーブル駆動）
  - 正常: `waiting` → `cancelled`
  - 正常: `completed` → `cancelled`
  - 異常: `cancelled` → 409
  - 異常: 存在しない id → 404
  - 異常: 権限なし → 403（既存 `RequirePermission` で担保）
- [ ] 既存 `accounting_handler_test.go:100` の DeleteAccounting テストを CancelAccounting に置き換え
- [ ] `CountItemsByBillingID` が他から呼ばれていなければ削除（`grep -rn "CountItemsByBillingID" backend/` で確認）
- [ ] `go test ./... -race` パス
- [ ] `golangci-lint run` パス
