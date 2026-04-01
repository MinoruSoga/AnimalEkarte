---
status: open
priority: high
created: 2026-04-01
---

# BUG-109: 物販品目削除（FK依存チェックなし）→ 204 で請求・在庫データ孤立

## 概要

`DELETE /api/v1/masters/merchandise-items/:id` が `treatment_items` / `invoice_items` への参照チェックをせず 204 を返す。

## 再現手順

1. 物販品目（例: ID=1）を使用している請求または処置アイテムが存在する状態で
2. `DELETE /api/v1/masters/merchandise-items/1` を実行
3. 結果: **204 No Content** が返され、データベースに孤立した参照が残される

**期待する動作:**
- `treatment_items` or `invoice_items` に当該 ID が存在する場合 → **409 Conflict** を返す

## 影響

- 物販品目の削除により、関連する請求・処置データが孤立する
- データ整合性が破壊される
- マスタ管理品目のテーブル整合性違反

## 実装場所

### バックエンド

**`backend/internal/service/merchandise_item_service.go`**
- `Delete()` メソッドにFK依存チェック追加

**`backend/internal/repository/merchandise_item_repository.go`**
- `CountTreatmentItemsByMerchandiseID(ctx, id)` メソッド追加
- `CountInvoiceItemsByMerchandiseID(ctx, id)` メソッド追加

### 実装コード例

```go
// repository/merchandise_item_repository.go
func (r *MerchandiseItemRepository) CountTreatmentItemsByMerchandiseID(ctx context.Context, id uint64) (int64, error) {
    var count int64
    return count, r.db.WithContext(ctx).Model(&model.TreatmentItem{}).
        Where("merchandise_item_id = ?", id).
        Count(&count).Error
}

func (r *MerchandiseItemRepository) CountInvoiceItemsByMerchandiseID(ctx context.Context, id uint64) (int64, error) {
    var count int64
    return count, r.db.WithContext(ctx).Model(&model.InvoiceItem{}).
        Where("merchandise_item_id = ?", id).
        Count(&count).Error
}

// service/merchandise_item_service.go
func (s *MerchandiseItemService) Delete(ctx context.Context, id uint64) error {
    // FK依存チェック
    treatmentCount, err := s.repo.CountTreatmentItemsByMerchandiseID(ctx, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to check treatment items dependency")
    }
    if treatmentCount > 0 {
        return apperrors.ErrConflict
    }

    invoiceCount, err := s.repo.CountInvoiceItemsByMerchandiseID(ctx, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to check invoice items dependency")
    }
    if invoiceCount > 0 {
        return apperrors.ErrConflict
    }

    // 削除実行
    return s.repo.Delete(ctx, id)
}
```

## テスト

- 依存データなし → 200 OK (削除成功)
- 処置アイテム参照あり → 409 Conflict
- 請求アイテム参照あり → 409 Conflict

## 優先度

**High** — データ整合性破壊（FK違反）

## 検証済み

- ステージング (stg.noah-karte.com) で 2026-04-01 確認: `DELETE /api/v1/masters/merchandise-items/1 → 204` で未修正確認
- `backend/internal/service/merchandise_item_service.go:Delete()` に FK チェックコードなし（コードレビュー確認済み）

## 関連

- BUG-030（依存チェックなし - 一般パターン）
- BUG-103（ケージ削除FK check - 修正済み ✅）
- BUG-107（処置マスタ削除FK check - 修正済み ✅）
- BUG-117（主訴カテゴリ削除FK check - 修正済み ✅）
- BUG-118（全マスタ重複名称UNIQUE制約 - 修正済み ✅）
