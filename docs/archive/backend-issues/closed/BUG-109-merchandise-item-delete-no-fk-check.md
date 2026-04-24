---
status: open
---

# BE: BUG-109 物販品目削除（FK依存チェックなし）→ 204 で請求・在庫データ孤立

## 概要

`DELETE /api/v1/masters/merchandise-items/:id` が `treatment_items` / `invoice_items` への参照チェックをせず 204 を返す。

## 再現手順

```
DELETE /api/v1/masters/merchandise-items/1 → 204 No Content
```

## 期待する動作

- `treatment_items` or `invoice_items` に当該 ID が存在する場合 → **409 Conflict**

## 実装場所

- `backend/internal/service/merchandise_item_service.go` — Delete() に依存チェック追加
- `backend/internal/repository/merchandise_item_repository.go` — `CountTreatmentItemsByMerchandiseID(ctx, id)` 追加

```go
count, err := s.repo.CountTreatmentItemsByMerchandiseID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check merchandise dependencies")
}
if count > 0 {
    return apperrors.ErrConflict
}
```

## 優先度

High（データ整合性破壊）

## ステータス確認

- ステージング (stg.noah-karte.com) で 2026-04-01 確認: `DELETE /api/v1/masters/merchandise-items/1 → 204` で未修正確認
- `backend/internal/service/merchandise_item_service.go:Delete()` に FK チェックコードなし（コードレビュー確認済み）
- チケットが誤って `closed/` に移動されていたため `open/` に戻した

## 関連

- BUG-030, BUG-103（同種パターン、BUG-103は修正済み）
- `docs/tasks/open/crash/BUG-109_merchandise-item-delete-no-fk-check.md`
