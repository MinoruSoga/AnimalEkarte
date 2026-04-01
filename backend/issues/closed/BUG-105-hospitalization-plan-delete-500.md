---
status: closed
---

# BE: BUG-105 入院プラン削除（FK制約クラッシュ）→ 500 Internal Server Error

## 概要

`DELETE /api/v1/masters/hospitalization-plans/:id` が FK 制約違反を適切にハンドリングせず 500 を返す。
サービス種別・スタッフでは修正済みの同種バグが入院プランハンドラで未修正。

## 再現手順

```
DELETE /api/v1/masters/hospitalization-plans/1 → 500 Internal Server Error
```

## 期待する動作

- `hospitalization_records.plan_id = :id` が存在する場合 → **409 Conflict**（500 ではなく）
- エラー: 「このデータは他のレコードに使用されています」

## 実装場所

- `backend/internal/service/hospitalization_plan_service.go` — Delete() に依存チェック追加
- `backend/internal/repository/hospitalization_plan_repository.go` — `CountRecordsByPlanID(ctx, id)` 追加

```go
count, err := s.repo.CountRecordsByPlanID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check plan dependencies")
}
if count > 0 {
    return apperrors.ErrConflict
}
```

## 優先度

High（500 クラッシュ相当）

## 関連

- BUG-030（サービス種別・スタッフは修正済み）
- BUG-108（薬剤削除も同種 500 クラッシュ）
- `docs/tasks/open/crash/BUG-105_hospitalization-plan-delete-500.md`
