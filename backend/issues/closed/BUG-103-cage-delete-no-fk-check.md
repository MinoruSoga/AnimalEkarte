---
status: closed
---

# BE: BUG-103 ケージ削除（FK依存チェックなし）→ 204 で孤立参照

## 概要

`DELETE /api/v1/masters/cages/:id` が入院データの存在チェックをせず 204 を返す。
`hospitalization_records.cage_id` が孤立しデータ整合性が破壊される。

## 再現手順

```
DELETE /api/v1/masters/cages/1 → 204 No Content（入院データ存在時も削除成功）
```

## 期待する動作

- `hospitalization_records.cage_id = :id` が存在する場合 → **409 Conflict**
- エラー: 「このデータは他のレコードに使用されているため削除できません」

## 実装場所

- `backend/internal/service/cage_service.go` — Delete() に依存チェック追加
- `backend/internal/repository/cage_repository.go` — `CountHospitalizationsByCageID(ctx, id)` 追加

```go
// cage_service.go Delete() 内
count, err := s.repo.CountHospitalizationsByCageID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check cage dependencies")
}
if count > 0 {
    return apperrors.ErrConflict
}
```

## 優先度

High（データ整合性破壊）

## 関連

- BUG-030（同種パターン。サービス種別・スタッフは修正済み）
- `docs/tasks/open/crash/BUG-103_cage-delete-no-fk-check.md`
