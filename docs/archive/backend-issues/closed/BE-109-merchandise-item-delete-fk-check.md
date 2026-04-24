---
id: BE-109
status: open
priority: high
created: 2026-04-01
parent_task: BUG-109
---

# BE-109: 物販品目削除 — FK依存チェック実装

## 概要

`DELETE /api/v1/masters/merchandise-items/:id` が `treatment_items` / `invoice_items` への FK参照チェックをせず 204 を返す。

## 影響

- 物販品目を削除すると、関連する請求・処置データが孤立
- データ整合性破壊（FK違反）
- **ローカル・ステージング両環境で確認済み** (2026-04-01)

## 実装対象ファイル

### 1. repository/merchandise_item_repository.go

以下の2メソッドを追加：

```go
// CountTreatmentItemsByMerchandiseID はこのマスタを参照する処置アイテム数を返す
func (r *MerchandiseItemRepository) CountTreatmentItemsByMerchandiseID(ctx context.Context, id uint64) (int64, error) {
    var count int64
    return count, r.db.WithContext(ctx).Model(&model.TreatmentItem{}).
        Where("merchandise_item_id = ?", id).
        Count(&count).Error
}

// CountInvoiceItemsByMerchandiseID はこのマスタを参照する請求アイテム数を返す
func (r *MerchandiseItemRepository) CountInvoiceItemsByMerchandiseID(ctx context.Context, id uint64) (int64, error) {
    var count int64
    return count, r.db.WithContext(ctx).Model(&model.InvoiceItem{}).
        Where("merchandise_item_id = ?", id).
        Count(&count).Error
}
```

### 2. service/merchandise_item_service.go

`Delete()` メソッドを以下に修正：

```go
func (s *merchandiseItemService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to get merchandise item")
	}

	// FK依存チェック: treatment_items
	treatmentCount, err := s.repo.CountTreatmentItemsByMerchandiseID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check treatment items dependency")
	}
	if treatmentCount > 0 {
		return apperrors.WrapConflict("この物販品目は処置で使用中のため削除できません")
	}

	// FK依存チェック: invoice_items
	invoiceCount, err := s.repo.CountInvoiceItemsByMerchandiseID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check invoice items dependency")
	}
	if invoiceCount > 0 {
		return apperrors.WrapConflict("この物販品目は請求で使用中のため削除できません")
	}

	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete merchandise item")
	}
	slog.InfoContext(ctx, "merchandise item deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("merchandise_item_id", id),
	)
	return nil
}
```

## テストケース

- ✅ 依存データなし → 204 No Content (削除成功)
- ✅ 処置アイテム参照あり → 409 Conflict (「この物販品目は処置で使用中のため削除できません」)
- ✅ 請求アイテム参照あり → 409 Conflict (「この物販品目は請求で使用中のため削除できません」)

## 参考実装

同じパターンで実装済み：
- **BUG-103 (ケージ)**: `cage_service.Delete()` — `hospitalizationRepo.ExistsByCageID()`
- **BUG-107 (処置)**: `procedure_service.Delete()` — `repo.CountUsageByProcedureID()`
- **BUG-112 (役職)**: `job_title_service.Delete()` — `repo.CountStaffsByJobTitleID()`

## 優先度

**High** — データ整合性破壊（FK違反）

## 検証済み実装状況

- **2026-04-01**: ローカル環境で `merchandise_item_service.go:Delete()` にFK依存チェックがないことを確認
- 他の4マスタ（ケージ・役職・薬剤・処置）には既に実装済み

## 関連

- Parent Task: BUG-109
- Related: BUG-103, BUG-107, BUG-112 (同パターン、既修正)
