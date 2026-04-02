---
status: closed
---

# BE: BUG-108 薬剤削除（FK制約クラッシュ）→ 500 Internal Server Error

## 概要

`DELETE /api/v1/masters/medicines/:id` が FK 制約違反を適切にハンドリングせず 500 を返す。
BUG-105（入院プラン）と同種のクラッシュ。

## 再現手順

```
DELETE /api/v1/masters/medicines/1 → 500 Internal Server Error
```

## 期待する動作

- `treatment_items.medicine_id = :id` が存在する場合 → **409 Conflict**
- エラー: 「このデータは他のレコードに使用されているため削除できません」

## 実装場所

- `backend/internal/service/medicine_service.go` — Delete() に依存チェック追加
- `backend/internal/repository/medicine_repository.go` — `CountTreatmentItemsByMedicineID(ctx, id)` 追加

```go
count, err := s.repo.CountTreatmentItemsByMedicineID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check medicine dependencies")
}
if count > 0 {
    return apperrors.ErrConflict
}
```

## 優先度

High（500 クラッシュ相当）

## 関連

- BUG-030, BUG-105（同種パターン）
- `docs/tasks/open/crash/BUG-108_medicine-delete-500.md`
